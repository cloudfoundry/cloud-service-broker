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

output name { value = aws_db_instance.db_instance.name }
output hostname { value = aws_db_instance.db_instance.address }
output port { value = local.ports[var.engine] }
output username { value = aws_db_instance.db_instance.username }
output password { value = aws_db_instance.db_instance.password }
output use_tls { value = var.use_tls }
output status {value = format("created db %s (id: %s) on server %s URL: https://%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s;is-cluster=false",
                               aws_db_instance.db_instance.name,
                               aws_db_instance.db_instance.id,
                               aws_db_instance.db_instance.address,
                               var.region,
                               var.region,
                               aws_db_instance.db_instance.id)}