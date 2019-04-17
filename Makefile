
clean/dev:
	kubectl delete --all deployments,canaries,pods,virtualservices -n krane

prepare/dev:
	kubectl apply -f ./example/00-canary-policy.yaml
	kubectl apply -f ./example/01-gw.yaml

deploy/dev:
	kubectl apply -f ./example/01-nginx-deployment.yaml
	kubectl apply -f ./example/01-vs.yaml
	kubectl apply -f ./example/02-nginx-canary.yaml