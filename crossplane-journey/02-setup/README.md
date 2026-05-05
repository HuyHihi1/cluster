# Bài 2: Cài đặt và Cấu hình

Để bắt đầu, bạn cần một Kubernetes cluster (Kind, Minikube, hoặc Docker Desktop).

## 1. Cài đặt Crossplane qua Helm
```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm install crossplane \
  --namespace crossplane-system \
  --create-namespace \
  crossplane-stable/crossplane
```

## 2. Kiểm tra trạng thái
```bash
kubectl get pods -n crossplane-system
```

## 3. Cài đặt Crossplane CLI (Tùy chọn nhưng khuyến khích)
Công cụ này giúp bạn quản lý các package dễ dàng hơn.
```bash
curl -sL "https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh" | sh
sudo mv kubectl-crossplane /usr/local/bin
```

## 4. Provider đầu tiên
Trong khóa học này, chúng ta sẽ dùng `provider-kubernetes` để thực hành (vì nó không tốn tiền và không cần tài khoản Cloud).

### Cài đặt Provider:
Xem file `exercises/01-install-provider.yaml`.

---
**Tiếp theo:** [Bài 3: Managed Resources](../03-managed-resources/README.md)
