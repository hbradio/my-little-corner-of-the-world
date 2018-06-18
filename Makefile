master: lambdas site-master

preview: lambdas site-preview

lambdas:
	mkdir -p lambdas
	go get ./...
	go build -o lambdas/scrape-instagram ./src/github.com/aninternetof/my-little-corner-of-the-world/scrape-instagram/...

site-master:
	yarn build

site-preview:
	yarn build-preview

clean-lambdas:
	rm -rf lambdas
