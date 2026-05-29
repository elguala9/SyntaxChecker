variable "region" {
  type    = string
  default = "eu-west-1"
}

resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"
  count         = 2

  tags = {
    Name = "web-server"
    Env  = "prod"
  }
}
