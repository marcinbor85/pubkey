#!/bin/bash

docker run -d -p 8080:8080 -v /storage:/storage --name pubkey pubkey
