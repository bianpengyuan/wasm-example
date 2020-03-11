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

#include <list>
#include <string>
#include <unordered_map>

#include "example/opa/opa_payload.pb.h"

namespace Opa {

// LRU cache for OPA policy check result.
class ResultCache {
public:
  bool check(const istio::wasm::example::opa::OpaPayload &payload,
             uint64_t &hash, bool &allowed, uint64_t timestamp);
  void add(const uint64_t hash, bool result, uint64_t timestamp);

private:
  void use(const uint64_t hash);

  std::unordered_map<
      uint64_t /* hash */,
      std::pair<bool /* result */, uint64_t /* insertion timestamp */>>
      result_cache_;
  std::list<uint64_t /* hash */> recent_;
  std::unordered_map<uint64_t /* hash */, std::list<uint64_t>::iterator> pos_;
};

} // namespace Opa
