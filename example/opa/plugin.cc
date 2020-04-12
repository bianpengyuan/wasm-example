/* Copyright 2020 Istio Authors. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "example/opa/plugin.h"

#include <google/protobuf/util/json_util.h>

#include "example/opa/opa_payload.pb.h"

// TODO: avoid using protobuf.
using google::protobuf::util::JsonParseOptions;
using google::protobuf::util::MessageToJsonString;
using google::protobuf::util::Status;

namespace {

inline bool
unmarshalConfig(istio::wasm::example::opa::OpaPluginConfig *opa_config) {
  WasmDataPtr configuration = getConfiguration();
  JsonParseOptions json_options;
  json_options.ignore_unknown_fields = true;
  Status status =
      JsonStringToMessage(configuration->toString(), opa_config, json_options);
  if (status != Status::OK) {
    LOG_WARN("Cannot parse OPA plugin configuration JSON string " +
             configuration->toString() + ", " + status.message().ToString());
    return false;
  }
  return true;
}

inline bool
unmarshalOpaResponse(const std::string &body,
                     istio::wasm::example::opa::OpaResponse *opa_response) {
  WasmDataPtr configuration = getConfiguration();
  JsonParseOptions json_options;
  json_options.ignore_unknown_fields = true;
  Status status = JsonStringToMessage(body, opa_response, json_options);
  if (status != Status::OK) {
    LOG_WARN("Cannot parse OPA Response JSON string " + body + ", " +
             status.message().ToString());
    return false;
  }
  return true;
}

} // namespace

namespace Opa {

bool OpaPluginRootContext::onStart(size_t) { return true; }

bool OpaPluginRootContext::validateConfiguration(
    size_t /* configuration_size */) {
  // Try parsing the configuration.
  istio::wasm::example::opa::OpaPluginConfig unused_config;
  return unmarshalConfig(&unused_config);
}

bool OpaPluginRootContext::onConfigure(size_t /* configuration_size */) {
  if (!unmarshalConfig(&config_)) {
    return false;
  }

  Metric cache_count(MetricType::Counter, "policy_cache_count",
                     {MetricTag{"wasm_filter", MetricTag::TagType::String},
                      MetricTag{"cache", MetricTag::TagType::String}});
  cache_hits_ = cache_count.resolve("opa_filter", "hit");
  cache_misses_ = cache_count.resolve("opa_filter", "miss");

  return true;
}

FilterHeadersStatus OpaPluginStreamContext::onRequestHeaders(uint32_t) {
  auto *root_context = getRootContext();
  istio::wasm::example::opa::OpaPayload payload;

  // Fill in payload proto.
  auto input = payload.mutable_input();
  *input->mutable_source_principal() = sourcePrincipal();
  std::string unused_dst_svc;
  destinationService(input->mutable_destination_service(), &unused_dst_svc);
  getValue({"request", "method"}, input->mutable_request_operation());
  getValue({"request", "url_path"}, input->mutable_request_url_path());

  uint64_t payload_hash = 0;
  bool allowed = false;
  bool cache_hit = root_context->checkCache(payload, payload_hash, allowed);
  if (cache_hit && allowed) {
    return FilterHeadersStatus::Continue;
  }
  if (cache_hit && !allowed) {
    sendLocalResponse(403, "OPA policy check denied", "", {});
    return FilterHeadersStatus::StopIteration;
  }

  // Convert payload proto to json string and send it to OPA server.
  std::string json_payload;
  if (MessageToJsonString(payload, &json_payload) != Status::OK) {
    // TODO add direct response
    LOG_WARN("cannot marshal OPA json payload");
    sendLocalResponse(500, "OPA policy check failed", "", {});
    return FilterHeadersStatus::StopIteration;
  }
  LOG_INFO("!!!!!!!!!!!!!! json payload is " + json_payload);

  // Construct http call to OPA server.
  HeaderStringPairs headers;
  HeaderStringPairs trailers;
  headers.emplace_back("content-type", "application/json");
  headers.emplace_back(":path", "/v1/data/test/allow");
  headers.emplace_back(":method", "POST");
  headers.emplace_back(":authority", root_context->opaServiceHost());

  // Get id of current context, which will be used in http callback.
  auto context_id = id();
  auto call_result = root_context->httpCall(
      root_context->opaClusterName(), headers, json_payload, trailers,
      /* timeout_milliseconds= */ 5000,
      [this, context_id, payload_hash](uint32_t, size_t body_size, uint32_t) {
        // Callback is triggered inside root context. setEffectiveContext
        // swtich the background context from root context to the current
        // stream context.
        getContext(context_id)->setEffectiveContext();
        auto body =
            getBufferBytes(BufferType::HttpCallResponseBody, 0, body_size);
        istio::wasm::example::opa::OpaResponse opa_response;
        LOG_INFO("!!!!!!!!!!!!!! body is " + body->toString());
        if (!unmarshalOpaResponse(body->toString(), &opa_response)) {
          // direct response.
          LOG_WARN("cannot unmarshal OPA response");
          sendLocalResponse(500, "OPA policy check failed", "", {});
          return;
        }
        this->getRootContext()->addCache(payload_hash, opa_response.result());
        if (!opa_response.result()) {
          // denied, send direct response.
          sendLocalResponse(403, "OPA policy check denied", "", {});
          return;
        }
        // allowed, continue request.
        continueRequest();
      });

  if (call_result != WasmResult::Ok) {
    LOG_WARN("cannot make call to OPA policy server");
    sendLocalResponse(500, "OPA policy check failed", "", {});
    return FilterHeadersStatus::StopIteration;
  }

  return FilterHeadersStatus::StopIteration;
}

} // namespace Opa
