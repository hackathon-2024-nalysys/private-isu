.PHONY: init bench connect proxy alp
init: webapp/sql/dump.sql.bz2 benchmarker/userdata/img

WORKER := isu-worker.grainrigi.net

webapp/sql/dump.sql.bz2:
	cd webapp/sql && \
	curl -L -O https://github.com/catatsuy/private-isu/releases/download/img/dump.sql.bz2

benchmarker/userdata/img.zip:
	cd benchmarker/userdata && \
	curl -L -O https://github.com/catatsuy/private-isu/releases/download/img/img.zip

benchmarker/userdata/img: benchmarker/userdata/img.zip
	cd benchmarker/userdata && \
	unzip -qq -o img.zip

bench:
	ssh -i ~/.ssh/private-isu.pem ubuntu@$(WORKER) "sudo killall -USR2 app && sudo rm /var/log/nginx/access.log && sudo systemctl restart nginx" && \
	./bench.sh

connect:
	ssh -i ~/.ssh/private-isu.pem isucon@$(WORKER)

proxy:
	ssh -i ~/.ssh/private-isu.pem -L 3306:localhost:3306 -L 11211:localhost:11211 ubuntu@$(WORKER)

dump:
	killall -USR1 app && journalctl -u isu-go.service -n 10000

reset:
	killall -USR2 app 

alp:
	sudo cat /var/log/nginx/access.log | alp ltsv --sort sum -m "posts/[0-9]+,/@\w+,/image/\d+" -o count,method,uri,min,avg,max,sum | less
