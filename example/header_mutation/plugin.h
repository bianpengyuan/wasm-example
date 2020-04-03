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

#include "example/header_mutation/config.pb.h"
#include "example/header_mutation/header_mutation.pb.h"

namespace HeaderMutation {

// HeaderMutationRootContext models a VM wide context.
class HeaderMutationRootContext
    : public Istio::Extension::ExtensionRootContext {
public:
  HeaderMutationRootContext(uint32_t id, StringView root_id)
      : ExtensionRootContext(id, root_id) {}

  bool onStart(size_t /* vm_configuration_size */) override;
  bool onConfigure(size_t /* configuration_size */) override;

  bool callHeaderMutation(
      const istio::wasm::example::header_mutation::HeaderMutationRequest &req,
      const uint32_t context_id);

private:
  istio::wasm::example::header_mutation::HeaderMutationPluginConfig config_;

  // Serialized string of header mutation service
  std::string grpc_service_string_;
};

// HeaderMutationContext models every HTTP request stream.
class HeaderMutationContext : public Istio::Extension::ExtensionStreamContext {
public:
  HeaderMutationContext(uint32_t id, ::RootContext *root)
      : ExtensionStreamContext(id, root) {}

  FilterHeadersStatus onRequestHeaders(uint32_t) override;

private:
  HeaderMutationRootContext *getRootContext() {
    auto *root = this->root();
    return dynamic_cast<HeaderMutationRootContext *>(root);
  }
};

} // namespace HeaderMutation

static RegisterContextFactory register_HeaderMutationContext(
    CONTEXT_FACTORY(HeaderMutation::HeaderMutationContext),
    ROOT_FACTORY(HeaderMutation::HeaderMutationRootContext));