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

void addStringLabel(rapidjson::Writer<rapidjson::StringBuffer>& writer, const std::string& key, const std::string& val) {
  writer.String(key);
  writer.String(val);
}

void addNumericLabel(rapidjson::Writer<rapidjson::StringBuffer>& writer, const std::string& key, double val) {
  writer.String(key);
  writer.Double(val);
}

}

bool LoggingRootContext::onStart(size_t) { return true; }

bool LoggingRootContext::onConfigure(size_t /* configuration_size */) {
  WasmDataPtr configuration = getConfiguration();

  rapidjson::Document d;
  d.Parse(configuration->toString());
  
  rapidjson::Value& config = d["logging_service"];
  logging_service_ = config.GetString();

  // Start Timer
  return true;
}

void LoggingRootContext::addLogEntry(LoggingContext* stream) {
  auto &src_info = stream->getSourceNodeInfo();
  auto &dst_info = stream->getDestinationNodeInfo();
  addStringLabel(log_writer_, "destination_name", dst_info.name());
  addStringLabel(log_writer_, "source_name", src_info.name());
  addNumericLabel(log_writer_, "latency", stream->duration());
  addStringLabel(log_writer_, "destination_service", stream->destinationServiceHost());
}

void LoggingContext::onLog() {
  getRootContext()->addLogEntry(this);
}

} // namespace Logging
