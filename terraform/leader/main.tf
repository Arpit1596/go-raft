

provider "aws" {
  region = "eu-north-1"
}

resource "aws_security_group" "sg_raft" {
  name = "sg_raft"
  description = "sg_raft"
  vpc_id = "vpc-0a39d0f477a1ea191"
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    to_port = 22
    protocol = "tcp"
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 50081
    to_port = 50081
    protocol = "tcp"
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 50082
    to_port = 50082
    protocol = "tcp"
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    to_port = 0
    protocol = "-1"
  }
  tags = {
    "Name" = "sg_raft"
  }
}


resource "aws_instance" "leader-election" {
  instance_type = "t3.small"
  ami = "ami-000e50175c5f86214" 
  subnet_id = "subnet-03219772e7aa25ff4"
  security_groups = [aws_security_group.sg_raft.id]
  key_name = "keypair"
  private_ip = "10.0.1.10"
  user_data = file("install.sh")
  monitoring = true
  tags = {
    "Name" = "leader-election"
  }
}
