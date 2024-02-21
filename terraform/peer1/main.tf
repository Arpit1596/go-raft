

provider "aws" {
  region = "eu-north-1"
}

resource "aws_instance" "leader-election" {
  instance_type = "t3.small"
  ami = "ami-000e50175c5f86214" 
  subnet_id = "subnet-03219772e7aa25ff4"
  security_groups = ["sg-0cd07375abf06b983"]
  key_name = "keypair"
  private_ip = "10.0.1.11"
  user_data = file("install.sh")
  monitoring = true
  tags = {
    "Name" = "leader-election"
  }
}
