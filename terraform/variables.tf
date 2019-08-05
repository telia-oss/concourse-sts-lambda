variable "name_prefix" {
  type    = string
  default = "sts-lambda-example"
}

variable "region" {
  type    = string
  default = "eu-west-1"
}

variable "tags" {
  type = map(string)
  default = {
    environment = "dev"
    terraform   = "True"
  }
}
