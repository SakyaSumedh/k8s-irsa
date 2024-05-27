resource "aws_dynamodb_table" "dynamodb" {
  billing_mode = "PROVISIONED"
  hash_key     = "email"
  name         = "${local.env}-dynamodb"
  attribute {
    name = "name"
    type = "S"
  }
  attribute {
    name = "email"
    type = "S"
  }
  point_in_time_recovery {
    enabled = true
  }
  read_capacity  = 5
  write_capacity = 5
  global_secondary_index {
    name            = "name-Global-Index"
    hash_key        = "name"
    projection_type = "ALL"
    read_capacity   = 5
    write_capacity  = 5
  }
}