.PHONY: agent
agent:
	docker build agent -t quay.io/clwalsh/bootc-agent -f agent/Dockerfile
	docker push quay.io/clwalsh/bootc-agent

.PHONY: agent-arm64
agent-arm64:
	docker build agent -t quay.io/clwalsh/agent-arm64 --platform linux/arm64 -f agent/Dockerfile
	docker push quay.io/clwalsh/agent-arm6