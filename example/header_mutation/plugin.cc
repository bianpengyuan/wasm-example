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

#include "example/header_mutation/plugin.h"

#include <google/protobuf/util/json_util.h>

// TODO: avoid using protobuf.
using google::protobuf::util::JsonParseOptions;
using google::protobuf::util::MessageToJsonString;
using google::protobuf::util::Status;
using istio::wasm::example::header_mutation::HeaderMutationPluginConfig;
using istio::wasm::example::header_mutation::HeaderMutationRequest;
using istio::wasm::example::header_mutation::HeaderMutationResponse;

namespace {

inline bool
unmarshalConfig(HeaderMutationPluginConfig *header_mutation_config) {
  WasmDataPtr configuration = getConfiguration();
  JsonParseOptions json_options;
  json_options.ignore_unknown_fields = true;
  Status status = JsonStringToMessage(configuration->toString(),
                                      header_mutation_config, json_options);
  if (status != Status::OK) {
    LOG_WARN("Cannot parse header mutation plugin configuration JSON string " +
             configuration->toString() + ", " + status.message().ToString());
    return false;
  }
  return true;
}

inline void mutateHeader(const HeaderMutationResponse &resp) {
  const auto &header_mutation = resp.header_mutation();
  bool header_mutated = false;
  for (const auto &mutation : header_mutation) {
    auto val = getRequestHeader(mutation.first);
    if (val->size() == 0) {
      header_mutated = true;
      replaceRequestHeader(mutation.first, mutation.second);
    } else if (val->toString() != mutation.second) {
      header_mutated = true;
      addRequestHeader(mutation.first, mutation.second);
    }
  }
  if (header_mutated) {
    clearRouteCache();
  }
}

} // namespace

namespace HeaderMutation {

bool HeaderMutationRootContext::onStart(size_t) { return true; }

bool HeaderMutationRootContext::validateConfiguration(
    size_t /* configuration_size */) {
  // Try parsing the configuration.
  HeaderMutationPluginConfig unused_config;
  return unmarshalConfig(&unused_config);
}

bool HeaderMutationRootContext::onConfigure(size_t /* configuration_size */) {
  if (!unmarshalConfig(&config_)) {
    return false;
  }

  // Construct grpc_service for header mutation gRPC call.
  GrpcService grpc_service;
  grpc_service.mutable_google_grpc()->set_stat_prefix(
      "header_mutation_service");
  grpc_service.mutable_google_grpc()->set_target_uri(
      config_.header_mutation_service());
  grpc_service.SerializeToString(&grpc_service_string_);
  return true;
}

bool HeaderMutationRootContext::callHeaderMutation(
    const HeaderMutationRequest &req, const uint32_t context_id) {
  grpcSimpleCall(
      grpc_service_string_, /* service_name= */
      "istio.wasm.example.header_mutation.HeaderMutationService",
      /* service_method= */ "GetHeaderMutation", req,
      /* timeout_milliseconds= */ 10000,
      /* success_callback= */
      [context_id](size_t body_size) {
        // Get response before switch to stream context.
        auto response =
            getBufferBytes(BufferType::GrpcReceiveBuffer, 0, body_size);
        getContext(context_id)->setEffectiveContext();
        mutateHeader(response->proto<HeaderMutationResponse>());
        continueRequest();
      },
      /* failure_callback= */
      [context_id](GrpcStatus status) {
        LOG_WARN("header mutation api call error: " +
                 std::to_string(static_cast<int>(status)) +
                 getStatus().second->toString());
        continueRequest();
      });
}

FilterHeadersStatus HeaderMutationContext::onRequestHeaders(uint32_t) {
  // Call out to gRPC server and get header to mutate/inject.
  // Get id of current context, which will be used in http callback.
  auto context_id = id();
  HeaderMutationRequest req;
  req.set_cookie(getRequestHeader("cookie")->toString());
  if (getRootContext()->callHeaderMutation(req, context_id)) {
    LOG_WARN("failed to call header mutation service");
    return FilterHeadersStatus::Continue;
  }
  return FilterHeadersStatus::StopIteration;
}

} // namespace HeaderMutation
