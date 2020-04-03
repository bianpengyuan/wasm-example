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

#include "example/opa/cache.h"
#include "istio/extension/extension.h"

const uint64_t DEFAULT_EXPIRATION = 10000000000; // nanoseconds = 10s
const uint64_t MAX_NUM_ENTRY = 1000;

namespace Opa {

namespace {

uint64_t computeHash(const istio::wasm::example::opa::OpaPayload &payload) {
  const uint64_t kMul = static_cast<uint64_t>(0x9ddfea08eb382d69);
  uint64_t h = 0;
  h += std::hash<std::string>()(payload.input().source_principal()) * kMul;
  h += std::hash<std::string>()(payload.input().destination_service()) * kMul;
  h += std::hash<std::string>()(payload.input().request_operation()) * kMul;
  h += std::hash<std::string>()(payload.input().request_url_path()) * kMul;
  return h;
}

} // namespace

bool ResultCache::check(const istio::wasm::example::opa::OpaPayload &param,
                        uint64_t &hash, bool &allowed, uint64_t timestamp) {
  hash = computeHash(param);
  auto iter = result_cache_.find(hash);
  if (iter == result_cache_.end()) {
    return false;
  }
  const auto &entry = iter->second;
  LOG_INFO("111111111 " + std::to_string(entry.second) + " " + std::to_string(DEFAULT_EXPIRATION) + " " + std::to_string(timestamp));
  if (entry.second + DEFAULT_EXPIRATION > timestamp) {
    use(hash);
    allowed = entry.first;
    return true;
  }
  auto recent_iter = pos_.find(hash);
  recent_.erase(recent_iter->second);
  pos_.erase(hash);
  result_cache_.erase(hash);
  return false;
}

void ResultCache::add(const uint64_t hash, bool result, uint64_t timestamp) {
  use(hash);
  result_cache_.emplace(hash, std::make_pair(result, timestamp));
}

void ResultCache::use(const uint64_t hash) {
  if (pos_.find(hash) != pos_.end()) {
    recent_.erase(pos_[hash]);
  } else if (recent_.size() >= MAX_NUM_ENTRY) {
    int old = recent_.back();
    recent_.pop_back();
    result_cache_.erase(old);
    pos_.erase(old);
  }
  recent_.push_front(hash);
  pos_[hash] = recent_.begin();
}

} // namespace Opa
