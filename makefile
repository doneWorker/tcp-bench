.PHONY: down restart restart-grafana build deploy forward build-client bench results

down:
	kubectl delete -f k8s/ --ignore-not-found
	kubectl delete -f jobs/ --ignore-not-found
	kubectl delete configmap grafana-dashboards --ignore-not-found

restart: down deploy

restart-grafana:
	kubectl rollout restart deployment/grafana

build:
	docker build -f go-std/Dockerfile  -t echo-std:v1  .
	docker build -f go-gnet/Dockerfile -t echo-gnet:v1 .

deploy:
	kubectl create configmap grafana-dashboards --from-file=dashboards/ --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f k8s/
	kubectl get pods

forward:
	kubectl port-forward svc/grafana 3000:3000

build-client:
	docker build -f client/Dockerfile -t echo-client:v1 .

bench:
	kubectl delete -f jobs/ --ignore-not-found
	kubectl apply -f jobs/

results:
	kubectl logs job/client-std
	kubectl logs job/client-gnet