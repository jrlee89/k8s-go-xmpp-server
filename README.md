# k8s go xmpp server
- Very minimal xmpp server k8s deployment.
- Work in progress.
- Inspired by https://github.com/mattn/go-xmpp/
- Provides only ANONYMOUS SASL & Server-Generated Resource Binding Stream Features.

### Deploy on minikube
```
podman login docker.io
podman build -t docker.io/<YOUR-REPO>/xmppsrv .
podman push docker.io/<YOUR-REPO>/xmppsrv
minikube start
kubectl create deployment xmppsrv --image=jrlee89/xmppsrv --port=5222 
kubectl expose deployment xmppsrv --type=NodePort --target-port=5222
```

### Test with xmpp client
```
git clone https://github.com/jrlee89/go-xmpp-example.git
```
- Set STARTTLS to false in xmpp.Options instance

Use tmux and run two instances of the example program.
```
$ minikube ip
192.168.39.206

$ kubectl get svc
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/xmppsrv      NodePort    10.104.36.105   <none>        5222:30540/TCP   11s

go run example.go -server=192.168.39.206:30540 -notls=true -debug=true
```
