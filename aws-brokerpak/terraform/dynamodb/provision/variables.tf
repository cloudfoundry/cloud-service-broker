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

variable labels { type = map }
variable aws_vpc_id { type = string }
variable instance_name { type = string }

variable billing_mode { type = string }
variable hash_key { type = string }
variable range_key { type = string }
variable tabel_name { type = string }

variable server_side_encryption_kms_key_arn { type = string }
variable attributes { type = list(map(string)) }
variable local_secondary_indexes { type = any }
variable global_secondary_indexes { type = any }

variable ttl_attribute_name { type = string }
variable ttl_enabled { type = bool }
variable stream_enabled { type = bool }
variable stream_view_type { type = string }
variable server_side_encryption_enabled { type = bool }

variable write_capacity { type = number }
variable read_capacity { type = number }