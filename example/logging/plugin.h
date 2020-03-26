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

class LoggingContext;

// LoggingRootContext models a VM wide context.
class LoggingRootContext
    : public Istio::Extension::ExtensionRootContext {
public:
  LoggingRootContext(uint32_t id, StringView root_id)
      : ExtensionRootContext(id, root_id) {
    log_writer_.Reset(log_buffer_);
  }

  bool onStart(size_t /* vm_configuration_size */) override;
  bool onConfigure(size_t /* configuration_size */) override;

  void addLogEntry(LoggingContext* stream_context);

private:
  std::string logging_service_;
  rapidjson::StringBuffer log_buffer_;
  rapidjson::Writer<rapidjson::StringBuffer> log_writer_;
};

// LoggingContext models every HTTP request stream.
class LoggingContext : public Istio::Extension::ExtensionStreamContext {
public:
  LoggingContext(uint32_t id, ::RootContext *root)
      : ExtensionStreamContext(id, root) {}

  void onLog() override;

private:
  LoggingRootContext *getRootContext() {
    auto *root = this->root();
    return dynamic_cast<LoggingRootContext *>(root);
  }
};

} // namespace Logging

static RegisterContextFactory register_LoggingContext(
    CONTEXT_FACTORY(Logging::LoggingContext),
    ROOT_FACTORY(Logging::LoggingRootContext));