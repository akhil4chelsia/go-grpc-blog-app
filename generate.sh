#!/bin/bash

protoc greet/greet.proto --go_out=plugins=grpc:.
protoc calculator/calculator.proto --go_out=plugins=grpc:.
protoc blog/blogpb/blog.proto --go_out=plugins=grpc:.