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

variable fog_name { type = string }
variable fog_resource_group { type = string }
variable db_name { type = string }

resource "null_resource" "run-failover" {

  provisioner "local-exec" {
    command = "sqlfailover ${var.fog_resource_group} ${var.fog_name} ${var.db_name} secondary" 
  }

  provisioner "local-exec" {
	when = destroy
    command = "sqlfailover ${var.fog_resource_group} ${var.fog_name} ${var.db_name} primary"
  }
}