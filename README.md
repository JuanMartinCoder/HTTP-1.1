# Protocol HTTPv1.1

## Description
This my own implementation of the HTTP protocol version 1.1 written in Golang over TCP.
The project takes care of the server implementation and the parser that takes care of reading the messages sent over TCP to build the HTTP Protocol itself.

## Resources used
- RFC 9112 -> (https://www.rfc-editor.org/rfc/rfc9112)
- RFC 9110 -> (https://www.rfc-editor.org/rfc/rfc9110)
- Go standard library for creating TCP servers

## Features made
- Capacity to implementing my own routing and handlers
- Use custom Headers and Trailers
- Set, Get, Replace and delete Headers
- Serve HTML, JSON, Images and Videos
- Also the server is able to use chunk encoding

