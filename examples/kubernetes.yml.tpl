---
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: default
spec:
  type: NodePort
  selector:
    app: myapp
  ports:
  - name: http
    port: 80
    targetPort: 80
    protocol: TCP
  - name: fpm
    port: 9000
    targetPort: 9000
    protocol: TCP
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: myapp
  namespace: default
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp-fpm
        resources:
          requests:
            cpu: "250m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        image: docker/myapp-fpm:master
        imagePullPolicy: Always
        readinessProbe:
          tcpSocket:
            port: 9000
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          tcpSocket:
            port: 9000
          initialDelaySeconds: 15
          periodSeconds: 20
        ports:
        - containerPort: 9000
        env:
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: TOPOLOGY
            value: preprod
          - name: APP_NAME
            value: myapp-fpm
      - name: myapp-nginx
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "300m"
            memory: "512Mi"
        image: docker/myapp-nginx:master
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: "/healthz"
            port: 80
          initialDelaySeconds: 30
          timeoutSeconds: 5
        ports:
        - containerPort: 80
        env:
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: TOPOLOGY
            value: preprod
          - name: APP_NAME
            value: myapp-nginx
---
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: myapp-scaler
spec:
  scaleTargetRef:
    kind: Deployment
    name: myapp
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 60
