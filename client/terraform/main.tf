provider "aws" {
  region = "eu-north-1"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support = true
  tags = {
    "Name" = "leader-election"
  }
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "private_subnet" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block = "10.0.1.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    "Name" = "private-subnet"
  }
}

resource "aws_subnet" "public_subnet" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block = "10.0.2.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    "Name" = "public-subnet"
  }
}

resource "aws_subnet" "public_subnet_1" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block = "10.0.3.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    "Name" = "public-subnet"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    "Name" = "igw"
  }
}

resource "aws_route_table" "publicrt" {
  vpc_id = aws_vpc.vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }
}

resource "aws_route_table_association" "publicrta" {
  subnet_id = aws_subnet.public_subnet.id
  route_table_id = aws_route_table.publicrt.id
}

resource "aws_route_table_association" "publicrta1" {
  subnet_id = aws_subnet.public_subnet_1.id
  route_table_id = aws_route_table.publicrt.id
}

resource "aws_eip" "eip" {
  vpc = true
}

resource "aws_nat_gateway" "ngw" {
  allocation_id = aws_eip.eip.id
  subnet_id = aws_subnet.public_subnet.id
  tags = {
    "Name" = "ngw"
  }
}

output "nat_gateway_ip" {
  value = aws_eip.eip.public_ip
}

resource "aws_route_table" "privatert" {
  vpc_id = aws_vpc.vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.ngw.id
  }
}

resource "aws_route_table_association" "privaterta" {
  subnet_id = aws_subnet.private_subnet.id
  route_table_id = aws_route_table.privatert.id
}

resource "aws_security_group" "sg" {
  name = "sg"
  description = "sg"
  vpc_id = aws_vpc.vpc.id
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 22
    to_port = 22
    protocol = "tcp"
  }

  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 8080
    to_port = 8080
    protocol = "tcp"
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    to_port = 0
    protocol = "-1"
  }
  tags = {
    "Name" = "sg"
  }
}

resource "aws_security_group" "alb_sg" {
  name = "alb_sg"
  description = "alb_sg"
  vpc_id = aws_vpc.vpc.id
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 80
    to_port = 80
    protocol = "tcp"
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port = 0
    to_port = 0
    protocol = "-1"
  }
  tags = {
    "Name" = "alb_sg"
  }
}

resource "aws_lb_target_group" "tg" {
  name       = "tg"
  port       = 8080
  protocol   = "HTTP"
  vpc_id     = aws_vpc.vpc.id
  slow_start = 0

  load_balancing_algorithm_type = "round_robin"

  stickiness {
    enabled = false
    type    = "lb_cookie"
  }

  health_check {
    enabled             = true
    port                = 8080
    interval            = 30
    protocol            = "HTTP"
    path                = "/health"
    matcher             = "200"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }
}

resource "aws_lb_target_group_attachment" "tga" {
  target_group_arn = aws_lb_target_group.tg.arn
  target_id        = aws_instance.leader-election-client.id
  port             = 8080
}

resource "aws_lb" "alb" {
  name               = "alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb_sg.id]

  subnets = [
    aws_subnet.public_subnet.id,
    aws_subnet.public_subnet_1.id
  ]
}

resource "aws_lb_listener" "albl" {
  load_balancer_arn = aws_lb.alb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.tg.arn
  }
}

resource "aws_key_pair" "keypair" {
  key_name   = "keypair"
  public_key = file("keys/mykeypair.pub")
}

resource "aws_instance" "leader-election-client" {
  instance_type = "t3.small"
  ami = "ami-000e50175c5f86214" 
  subnet_id = aws_subnet.private_subnet.id
  security_groups = [aws_security_group.sg.id]
  key_name = aws_key_pair.keypair.key_name
  private_ip = "10.0.1.13"
  user_data = file("install.sh")
  monitoring = true
  tags = {
    "Name" = "leader-election-client"
  }
}

resource "aws_instance" "jumpbox" {
  instance_type = "t3.small"
  ami = "ami-000e50175c5f86214" 
  subnet_id = aws_subnet.public_subnet.id
  security_groups = [aws_security_group.sg.id]
  key_name = aws_key_pair.keypair.key_name
  private_ip = "10.0.2.10"
  user_data = file("install.sh")
  monitoring = true
  tags = {
    "Name" = "jumpbox"
  }
}

resource "aws_eip" "jumphosteip" {
  instance = aws_instance.jumpbox.id
  vpc = true
}

output "jumphost_ip" {
  value = aws_eip.jumphosteip.public_ip
}



