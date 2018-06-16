master: lambdas site-master

preview: lambdas site-preview

lambdas:
	mkdir -p lambdas
	go get ./...
	go build -o lambdas/scrape-instagram ./...

site-master:
	yarn build

site-preview:
	yarn build-preview
