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

variable name {type = "string"}
variable labels {type = "map"}
variable credentials  { type = string }
variable project  { type = string }
variable role { type = string }

output name { value = var.name }
output credentials { value = var.credentials }
output project { value = var.project }
output role { value = var.role }
