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

#include "example/logging/plugin.h"

#include "proxy_wasm_intrinsics.pb.h"
#include "rapidjson/document.h"

namespace Logging {

namespace {

void startLogRequest(rapidjson::Writer<rapidjson::StringBuffer> &writer,
                   rapidjson::StringBuffer *buffer) {
  writer.Reset(*buffer);
  writer.StartArray();
}

void endLogRequest(rapidjson::Writer<rapidjson::StringBuffer> &writer) {
  writer.EndArray();
}

void addStringLabel(rapidjson::Writer<rapidjson::StringBuffer> &writer,
                    const std::string &key, const std::string &val) {
  writer.Key(key.c_str(), key.length(), true);
  writer.String(val);
}

void addNumericLabel(rapidjson::Writer<rapidjson::StringBuffer> &writer,
                     const std::string &key, double val) {
  writer.Key(key.c_str(), key.length(), true);
  writer.Double(val);
}

void addIntegerLabel(rapidjson::Writer<rapidjson::StringBuffer> &writer,
                     const std::string &key, int64_t val) {
  writer.Key(key.c_str(), key.length(), true);
  writer.Int64(val);
}

} // namespace

LoggingRootContext::LoggingRootContext(uint32_t id, StringView root_id)
    : ExtensionRootContext(id, root_id) {
  log_buffer_ = std::make_unique<rapidjson::StringBuffer>();
  log_writer_.Reset(*log_buffer_);
  log_writer_.StartArray();
  log_entry_count_ = 0;
}

bool LoggingRootContext::onStart(size_t) { 
  return true;
}

bool LoggingRootContext::onConfigure(size_t /* configuration_size */) {
  WasmDataPtr configuration = getConfiguration();

  rapidjson::Document d;
  d.Parse(configuration->toString());

  // Extract configuration for the logging filter
  if (d.HasMember("logging_service_cluster")) {
    rapidjson::Value &config = d["logging_service_cluster"];
    logging_service_cluster_ = config.GetString();
  }
  if (d.HasMember("logging_service_host")) {
    rapidjson::Value &config = d["logging_service_host"];
    logging_service_host_ = config.GetString();
  }

  // Start timer, which will trigger log report every 10s.
  proxy_set_tick_period_milliseconds(10000 /* milliseconds */);

  return true;
}

bool LoggingRootContext::onDone() {
  if (req_buffer_.empty() && log_entry_count_ == 0) {
    return true;
  }
  // Flush out all log entries
  flushLogBuffer();
  sendLogRequest(/* ondone */ true);
  return false;
}

void LoggingRootContext::onTick() {
  // Flush out all log entries
  flushLogBuffer();
  if (req_buffer_.empty()) {
    return;
  }
  sendLogRequest(/* ondone */ false);
}

void LoggingRootContext::addLogEntry(LoggingStreamContext *stream) {
  log_writer_.StartObject();
  // Add log labels
  // Workload labels
  std::string source_address, destination_address;
  getValue({"source", "address"}, &source_address);
  getValue({"destination", "address"}, &destination_address);
  addStringLabel(log_writer_, "source_name", stream->sourceName());
  addStringLabel(log_writer_, "source_namespace", stream->sourceNamespace());
  addStringLabel(log_writer_, "source_workload", stream->sourceWorkloadName());
  addStringLabel(log_writer_, "source_address", source_address);
  addStringLabel(log_writer_, "destination_name", stream->destinationName());
  addStringLabel(log_writer_, "destination_namespace", stream->destinationNamespace());
  addStringLabel(log_writer_, "destination_workload", stream->destinationWorkloadName());
  addStringLabel(log_writer_, "destination_address", destination_address);

  // Service labels
  std::string destination_service_name;
  std::string destination_service_host;
  stream->destinationService(&destination_service_host, &destination_service_name);
  addStringLabel(log_writer_, "destination_service_name", destination_service_name);
  addStringLabel(log_writer_, "destination_service_host", destination_service_host);

  // latency and timestamp
  int64_t timestamp;
  int64_t duration;
  getValue({"request", "time"}, &timestamp);
  getValue({"request", "duration"}, &duration);
  addIntegerLabel(log_writer_, "timestamp", timestamp);
  addNumericLabel(log_writer_, "latency", duration / 1000000);

  // request labels
  std::string url_path, url_host, request_id, referer, user_agent;
  int64_t response_code;
  getValue({"request", "url_path"}, &url_path);
  getValue({"request", "id"}, &request_id);
  getValue({"request", "host"}, &url_host);
  getValue({"response", "code"}, &response_code);
  getValue({"request", "referer"}, &referer);
  getValue({"request", "user_agent"}, &user_agent);
  addStringLabel(log_writer_, "url_path", url_path);
  addStringLabel(log_writer_, "response_flag", stream->responseFlag());
  addStringLabel(log_writer_, "source_principal", stream->sourcePrincipal());
  addStringLabel(log_writer_, "destination_principal", stream->destinationPrincipal());
  addStringLabel(log_writer_, "request_protocol", stream->requestProtocol());
  addIntegerLabel(log_writer_, "response_code", response_code);
  addStringLabel(log_writer_, "referer", referer);
  addStringLabel(log_writer_, "user_agent", user_agent);
  addStringLabel(log_writer_, "request_id", request_id);
  addStringLabel(log_writer_, "url_host", url_host);

  log_writer_.EndObject();
  log_entry_count_ += 1;
  if (log_entry_count_ >= 500) {
    flushLogBuffer();
  }
}

void LoggingRootContext::flushLogBuffer() {
  if (log_entry_count_ <= 0) {
    return;
  }
  endLogRequest(log_writer_);
  auto new_log_buffer = std::make_unique<rapidjson::StringBuffer>();
  log_buffer_.swap(new_log_buffer);
  req_buffer_.emplace_back(std::move(new_log_buffer));
  startLogRequest(log_writer_, log_buffer_.get());
  log_entry_count_ = 0;
}

void LoggingRootContext::sendLogRequest(bool ondone) {
  HeaderStringPairs headers;
  HeaderStringPairs trailers;
  headers.emplace_back("content-type", "application/json");
  headers.emplace_back(":path", "/");
  headers.emplace_back(":method", "POST");
  headers.emplace_back(":authority", logging_service_host_);
  uint32_t timeout_milliseconds = 10000;
  auto callback = [this, ondone](uint32_t, size_t, uint32_t) {
    LOG_INFO("received response from logging service.");
    in_flight_ondone_ -= 1;
    if (in_flight_ondone_ == 0 && ondone) {
      proxy_done();
    }
  };
  for (const auto &req : req_buffer_) {
    StringView body(req->GetString(), req->GetSize());
    auto call_result = httpCall(logging_service_cluster_, headers, body,
                                trailers, timeout_milliseconds, callback);

    if (call_result != WasmResult::Ok) {
      LOG_WARN("cannot make call to log server");
      break;
    }
    in_flight_ondone_ += 1;
  }
  req_buffer_.clear();
}

void LoggingStreamContext::onLog() { 
  getRootContext()->addLogEntry(this);
}

} // namespace Logging
