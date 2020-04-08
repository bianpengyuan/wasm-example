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

#pragma once

#include <map>

#include "istio/extension/extension.h"
#include "rapidjson/stringbuffer.h"
#include "rapidjson/writer.h"

namespace Logging {

class LoggingStreamContext;

class LoggingRootContext : public Istio::Extension::ExtensionRootContext {
public:
  LoggingRootContext(uint32_t id, StringView root_id);

  void onTick() override;

  bool onStart(size_t /* vm_configuration_size */) override;
  bool onConfigure(size_t /* configuration_size */) override;
  bool onDone() override;

  void addLogEntry(LoggingStreamContext *stream_context);
  void flushLogBuffer();
  void sendLogRequest(bool ondone);

private:
  std::string logging_service_cluster_;
  std::string logging_service_host_;
  std::unique_ptr<rapidjson::StringBuffer> log_buffer_;

  int log_entry_count_;
  rapidjson::Writer<rapidjson::StringBuffer> log_writer_;

  // Buffers requests to be sent.
  std::vector<std::unique_ptr<rapidjson::StringBuffer>> req_buffer_;

  int in_flight_ondone_ = 0;
};

class LoggingStreamContext : public Istio::Extension::ExtensionStreamContext {
public:
  LoggingStreamContext(uint32_t id, ::RootContext *root)
      : ExtensionStreamContext(id, root) {}

  void onLog() override;

private:
  LoggingRootContext *getRootContext() {
    auto *root = this->root();
    return dynamic_cast<LoggingRootContext *>(root);
  }
};

} // namespace Logging

static RegisterContextFactory
    register_LoggingStreamContext(CONTEXT_FACTORY(Logging::LoggingStreamContext),
                            ROOT_FACTORY(Logging::LoggingRootContext));