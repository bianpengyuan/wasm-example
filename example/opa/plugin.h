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

#include "example/opa/cache.h"
#include "example/opa/config.pb.h"
#include "istio/extension/extension.h"

namespace Opa {

// OpaPluginRootContext models a VM wide context. It has the same lifecycle as
// Wasm VM.
class OpaPluginRootContext : public Istio::Extension::ExtensionRootContext {
public:
  OpaPluginRootContext(uint32_t id, StringView root_id)
      : ExtensionRootContext(id, root_id) {}

  // onStart is called with Wasm VM configuration, it will only called once when
  // the Wasm VM is created. This will be the first call triggered inside the
  // Wasm VM.
  bool onStart(size_t /* vm_configuration_size */) override;

  // Validate OPA plugin configuration, which will cause an listener update to
  // be rejected.
  bool validateConfiguration(size_t /* configuration_size */) override;

  // Load OPA plugin configuration.
  bool onConfigure(size_t /* configuration_size */) override;

  const std::string &opaServiceHost() { return config_.opa_service_host(); }
  const std::string &opaClusterName() { return config_.opa_cluster_name(); }

  bool checkCache(const istio::wasm::example::opa::OpaPayload &payload,
                  uint64_t &hash, bool &allowed) {
    bool hit = cache_.check(payload, hash, allowed, getCurrentTimeNanoseconds());
    incrementMetric((hit ? cache_hits_ : cache_misses_), 1);
    return hit;
  }
  void addCache(const uint64_t hash, bool result) {
    cache_.add(hash, result, getCurrentTimeNanoseconds());
  }

private:
  istio::wasm::example::opa::OpaPluginConfig config_;
  ResultCache cache_;

  uint32_t cache_hits_;
  uint32_t cache_misses_;
};

// OpaPluginStreamContext models every HTTP request.
class OpaPluginStreamContext : public Istio::Extension::ExtensionStreamContext {
public:
  OpaPluginStreamContext(uint32_t id, ::RootContext *root)
      : ExtensionStreamContext(id, root) {}

  FilterHeadersStatus onRequestHeaders(uint32_t) override;

private:
  OpaPluginRootContext *getRootContext() {
    auto *root = this->root();
    return dynamic_cast<OpaPluginRootContext *>(root);
  }
};

} // namespace Opa

static RegisterContextFactory
    register_OpaPluginContext(CONTEXT_FACTORY(Opa::OpaPluginStreamContext),
                              ROOT_FACTORY(Opa::OpaPluginRootContext));