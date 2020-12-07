# Copyright 2020 Pivotal Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

variable cache_size { type = number }
variable redis_version { type = string }
variable instance_name { type = string }
variable labels { type = map }
variable aws_vpc_id { type = string }
variable node_type { type = string }
variable node_count { type = number }
variable elasticache_subnet_group { type = string }
variable vpc_security_group_ids { type = string }