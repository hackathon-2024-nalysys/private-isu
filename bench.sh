#!/bin/bash

ssh -i ~/.ssh/private-isu.pem ubuntu@isu-bench.grainrigi.net \
sudo -u isucon sh -c '"cd ~/private_isu.git/benchmarker && ./bin/benchmarker -u userdata -t http://isu-worker.grainrigi.net"'