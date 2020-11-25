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

resource "aws_elasticache_replication_group" "redis" {
  automatic_failover_enabled    = var.node_count > 1
  replication_group_id          = var.instance_name
  replication_group_description = format("%s redis", var.instance_name)
  node_type                     = local.node_type
  number_cache_clusters         = var.node_count
  parameter_group_name          = local.parameter_group_names[var.redis_version]
  port                          = local.port
  tags                          = var.labels
  security_group_ids            = [aws_security_group.sg.id]
  subnet_group_name             = aws_elasticache_subnet_group.subnet_group.name
  transit_encryption_enabled    = true
  auth_token                    = random_password.auth_token.result
}
