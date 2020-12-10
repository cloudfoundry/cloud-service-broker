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

variable cores { type = number }
variable instance_name { type = string }
variable db_name { type = string }
variable labels { type = map }
variable storage_gb { type = number }
variable publicly_accessible { type = bool }
variable multi_az { type = bool }
variable instance_class { type = string }
variable engine { type = string }
variable engine_version { type = string }
variable aws_vpc_id { type = string }
variable storage_autoscale { type = bool }
variable storage_autoscale_limit_gb { type = number }
variable storage_encrypted { type = bool }
variable parameter_group_name { type = string }
variable rds_subnet_group { type = string }
variable subsume { type = bool }
variable rds_vpc_security_group_ids { type = string }
variable allow_major_version_upgrade { type = bool }
variable auto_minor_version_upgrade { type = bool }
variable maintenance_window { type = string }
variable use_tls { type = bool }
