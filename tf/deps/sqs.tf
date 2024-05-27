resource "aws_sqs_queue" "queue" {
  name       = "${local.env}-sqs.fifo"
  fifo_queue = true
}