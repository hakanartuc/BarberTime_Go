# Kubernetes Kurulum Rehberi - BarberTime

Bu rehber, BarberTime uygulamasını Ubuntu Server üzerinde Kubernetes ve HPA (Horizontal Pod Autoscaler) ile kurmak için adım adım talimatlar içerir.

## Gereksinimler

- Ubuntu Server 20.04 veya üzeri
- Minimum 4 CPU, 8GB RAM (önerilen: 8 CPU, 16GB RAM)
- Root veya sudo yetkisi
- İnternet bağlantısı

## 1. Kubernetes Cluster Kurulumu

### 1.1. Sistem Güncellemeleri

```bash
sudo apt update && sudo apt upgrade -y
```

### 1.2. Container Runtime Kurulumu (containerd)

```bash
# containerd kurulumu
sudo apt install -y containerd
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo systemctl restart containerd
sudo systemctl enable containerd
```

### 1.3. Kubernetes Kurulumu

```bash
# Gerekli paketleri kur
sudo apt install -y apt-transport-https ca-certificates curl gpg

# Kubernetes için gerekli sistem paketleri
sudo apt install -y conntrack socat ebtables ipset

# Kubernetes için yeni repository (pkgs.k8s.io)
# GPG key'i ekle
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Repository ekle
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Kubernetes bileşenlerini kur
sudo apt update
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

**Not:** `conntrack`, `socat`, `ebtables` ve `ipset` paketleri Kubernetes'in preflight kontrolleri için gereklidir. Bu paketler olmadan `kubeadm init` komutu hata verecektir.

**Not:** Yukarıdaki komut Kubernetes v1.31 için hazırlanmıştır. Farklı bir versiyon kullanmak isterseniz, `v1.31` kısmını istediğiniz versiyonla değiştirin (örn: `v1.30`, `v1.32`).

### 1.4. Swap'ı Devre Dışı Bırak

```bash
sudo swapoff -a
sudo sed -i '/ swap / s/^/#/' /etc/fstab
```

**Neden Swap'ı Kapatıyoruz?**

Kubernetes, swap alanının kapatılmasını zorunlu kılar çünkü:

1. **Performans ve Tahmin Edilebilirlik**: Swap kullanımı, pod'ların kaynak kullanımını tahmin etmeyi zorlaştırır. Kubernetes, pod'ların bellek limitlerini doğru şekilde yönetemez.

2. **Kaynak Yönetimi**: Kubernetes, CPU ve bellek kaynaklarını pod'lara tahsis ederken swap'ı hesaba katmaz. Swap aktifken, bir pod bellek limitini aşsa bile sistem swap'a yazabilir ve bu durum performans sorunlarına yol açar.

3. **Kubelet Davranışı**: `kubelet` (Kubernetes'in node agent'ı) swap aktifken düzgün çalışmaz ve hata verir. Kubernetes v1.22'den itibaren swap desteği tamamen kaldırılmıştır.

4. **Pod Eviction (Çıkarma) Mekanizması**: Kubernetes, kaynak sıkıntısı olduğunda pod'ları otomatik olarak sonlandırır. Swap aktifken bu mekanizma doğru çalışmaz çünkü sistem swap kullanarak bellek basıncını gizler.

**Not**: Eğer sisteminizde swap kullanmak zorundaysanız (örneğin çok düşük RAM'li bir sistem), Kubernetes'in resmi dokümantasyonu swap'ı kapatmanızı önerir. Ancak bazı özel durumlarda `kubelet`'e `--fail-swap-on=false` parametresi ile çalıştırılabilir, ancak bu production ortamlarında önerilmez.

### 1.5. Kernel Parametrelerini Ayarla

```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

sudo modprobe overlay
sudo modprobe br_netfilter

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sudo sysctl --system
```

### 1.6. Kubernetes Cluster'ı Başlat (Master Node)

```bash
# Cluster'ı başlat
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# Kubeconfig'i kopyala
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

**Kubeconfig Sorun Giderme:**

Eğer `connection refused` hatası alıyorsanız:

```bash
# Kubeconfig'in varlığını kontrol et
ls -la $HOME/.kube/config

# Kubeconfig içeriğini kontrol et
cat $HOME/.kube/config

# Eğer yoksa veya bozuksa, yeniden kopyala
mkdir -p $HOME/.kube
sudo cp -f /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# KUBECONFIG environment variable'ını kontrol et
echo $KUBECONFIG

# Eğer farklı bir path gösteriyorsa, temizle veya doğru path'i ayarla
export KUBECONFIG=$HOME/.kube/config

# Test et
kubectl cluster-info
kubectl get nodes
```

**Pod Network CIDR Açıklaması:**

`--pod-network-cidr=10.244.0.0/16` parametresi, Kubernetes pod'larının kullanacağı IP adres aralığını belirler:

- **10.244.0.0/16**: Bu, 10.244.0.0 ile 10.244.255.255 arasındaki tüm IP adreslerini kapsar (toplam 65,536 IP adresi)
- **Neden bu IP aralığı?**: Flannel CNI plugin'inin varsayılan IP aralığıdır. Flannel kullanıyorsanız bu değeri kullanmalısınız.
- **Nasıl seçilir?**: 
  - Seçtiğiniz **CNI (Container Network Interface) plugin**'ine göre değişir
  - Mevcut ağınızla çakışmamalıdır (örneğin: 192.168.0.0/16, 172.16.0.0/12 gibi)
  - RFC 1918 özel IP aralıklarından biri olmalıdır

**Farklı CNI Plugin'leri için CIDR değerleri:**

| CNI Plugin | Önerilen CIDR | Not |
|------------|---------------|-----|
| Flannel | `10.244.0.0/16` | Varsayılan, en yaygın |
| Calico | `192.168.0.0/16` | Veya başka bir aralık |
| Weave | `10.32.0.0/12` | Veya başka bir aralık |
| Cilium | `10.0.0.0/8` | Veya başka bir aralık |

**Önemli Notlar:**

1. **Ağ çakışması kontrolü**: Bu IP aralığının mevcut ağınızla çakışmadığından emin olun:
   ```bash
   # Mevcut ağ yapılandırmasını kontrol et
   ip route show
   ```
   
2. **Flannel kullanıyorsanız**: `10.244.0.0/16` kullanın (yukarıdaki örnekte olduğu gibi)

3. **Farklı bir CNI kullanıyorsanız**: O plugin'in dokümantasyonuna göre CIDR değiştirin

4. **Multi-cluster ortamı**: Birden fazla cluster varsa, her birinin farklı CIDR kullanması gerekir

**Not:** `kubeadm init` komutunun çıktısında worker node'lar için birleştirme komutu görünecektir. Bunu kaydedin.

### 1.7. Pod Network Plugin Kurulumu (Flannel)

```bash
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```

**Flannel Nedir ve Ne İşe Yarar?**

Flannel, Kubernetes'te **pod'lar arası iletişimi sağlayan bir CNI (Container Network Interface) plugin**'idir. Kubernetes cluster'ında pod'ların birbirleriyle ve dış dünyayla iletişim kurabilmesi için mutlaka bir network plugin'e ihtiyaç vardır.

### Flannel'in Görevleri:

1. **Pod'lar Arası İletişim**:
   - Farklı node'larda çalışan pod'ların birbirleriyle iletişim kurmasını sağlar
   - Her pod'a benzersiz bir IP adresi atar
   - Pod'lar sanki aynı ağda gibi birbirlerine erişebilir

2. **Overlay Network (Örtü Ağı)**:
   - Fiziksel ağın üzerine sanal bir ağ katmanı oluşturur
   - Pod'ların IP adresleri fiziksel ağdan bağımsızdır
   - `10.244.0.0/16` gibi özel IP aralıkları kullanır

3. **Network Routing (Ağ Yönlendirme)**:
   - Pod'lardan gelen trafiği doğru node'a yönlendirir
   - VXLAN, UDP veya host-gw gibi farklı backend'ler kullanabilir
   - En yaygın kullanılan: VXLAN (Virtual Extensible LAN)

4. **IP Adresi Yönetimi**:
   - Her node'a bir alt ağ (subnet) tahsis eder
   - Örnek: Node 1 → `10.244.1.0/24`, Node 2 → `10.244.2.0/24`
   - Pod'lara bu alt ağlardan IP adresleri atar

### Flannel Nasıl Çalışır?

```
Pod A (Node 1)          Pod B (Node 2)
   10.244.1.5              10.244.2.10
       ↓                        ↓
   Flannel                  Flannel
   (VXLAN)                  (VXLAN)
       ↓                        ↓
   Fiziksel Ağ (192.168.1.0/24)
```

1. **Pod A** (Node 1'de, IP: 10.244.1.5) → **Pod B**'ye (Node 2'de, IP: 10.244.2.10) paket gönderir
2. Flannel, paketi VXLAN ile encapsulate eder (sarmalar)
3. Fiziksel ağ üzerinden Node 2'ye gönderir
4. Node 2'deki Flannel, paketi decapsulate eder (açar)
5. Paket Pod B'ye ulaşır

### Flannel vs Diğer CNI Plugin'leri:

| CNI Plugin | Özellikler | Kullanım |
|------------|------------|----------|
| **Flannel** | Basit, hafif, VXLAN/UDP | Küçük-orta ölçekli cluster'lar |
| **Calico** | Network Policy, BGP routing | Production, güvenlik önemli |
| **Weave** | Basit, otomatik keşif | Küçük cluster'lar |
| **Cilium** | eBPF tabanlı, yüksek performans | Büyük ölçekli, performans kritik |

### Flannel'in Avantajları:

1. **Basitlik**: Kurulumu ve yönetimi kolaydır
2. **Hafif**: Düşük kaynak kullanımı
3. **Stabil**: Uzun süredir kullanılan, olgun bir proje
4. **Varsayılan**: Birçok Kubernetes kurulumunda varsayılan seçim

### Flannel'in Dezavantajları:

1. **Network Policies**: Varsayılan olarak network policy desteği yoktur (Calico'da var)
2. **Performans**: Çok büyük cluster'larda performans sorunları olabilir
3. **Özellikler**: Calico veya Cilium kadar gelişmiş özelliklere sahip değildir

### Neden Flannel Kullanıyoruz?

- **Basit ve güvenilir**: Development ve küçük-orta ölçekli production için yeterlidir
- **Kolay kurulum**: Tek komutla kurulur
- **Yaygın kullanım**: Topluluk tarafından iyi bilinir ve desteklenir
- **BarberTime için yeterli**: Uygulamanızın ihtiyaçları için Flannel yeterlidir

### Alternatif Kullanım:

Eğer daha gelişmiş özellikler istiyorsanız (network policies, daha iyi performans), Calico veya Cilium kullanabilirsiniz. Ancak Flannel, çoğu kullanım senaryosu için yeterlidir.

**Önemli Not:** Flannel kurulmadan pod'lar birbirleriyle iletişim kuramaz. `kubeadm init` sonrası mutlaka bir CNI plugin kurulmalıdır.

### 1.8. Worker Node'ları Ekle (Opsiyonel - Multi-Node için)

Worker node'larda (master node hariç diğer sunucularda):

```bash
# Master node'dan aldığınız join komutunu çalıştırın
sudo kubeadm join <MASTER_IP>:6443 --token <TOKEN> --discovery-token-ca-cert-hash sha256:<HASH>
```

## 2. Metrics Server Kurulumu (HPA için gerekli)

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Metrics server'ın çalıştığını doğrula
kubectl get deployment metrics-server -n kube-system
```

Eğer metrics server başlamazsa, aşağıdaki patch'i uygulayın:

```bash
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'
```

## 3. Ingress Controller Kurulumu (Opsiyonel)

**Önce Cluster Durumunu Kontrol Edin:**

```bash
# Cluster durumunu kontrol et
kubectl cluster-info

# Node'ların durumunu kontrol et
kubectl get nodes

# API server'ın çalıştığını doğrula
kubectl get pods -n kube-system
```

Eğer `connection refused` hatası alıyorsanız:

1. **kubeconfig'in doğru olduğundan emin olun:**
   ```bash
   # kubeconfig'i kontrol et
   kubectl config view
   
   # Eğer yoksa, kopyalayın
   mkdir -p $HOME/.kube
   sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
   sudo chown $(id -u):$(id -g) $HOME/.kube/config
   ```

2. **API server'ın çalıştığını kontrol edin:**
   ```bash
   # API server pod'unu kontrol et
   kubectl get pods -n kube-system | grep kube-apiserver
   
   # Eğer çalışmıyorsa, logları kontrol edin
   kubectl logs -n kube-system kube-apiserver-<node-name>
   ```

3. **Network bağlantısını test edin:**
   ```bash
   # API server'a bağlantıyı test et
   curl -k https://localhost:6443/healthz
   ```

**Ingress Controller Kurulumu:**

```bash
# NGINX Ingress Controller kur
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Eğer validation hatası alırsanız (cluster henüz tam hazır değilse):
# kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml --validate=false

# Ingress controller'ın hazır olmasını bekle
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

**Sorun Giderme:**

Eğer `connection refused` hatası devam ediyorsa:

1. **Cluster'ın tamamen başlatıldığından emin olun:**
   ```bash
   # Tüm sistem pod'larının çalıştığını kontrol et
   kubectl get pods -n kube-system
   
   # Tüm pod'lar Running durumunda olmalı
   ```

2. **API server servisini kontrol edin:**
   ```bash
   # API server servisini kontrol et
   kubectl get svc -n default | grep kubernetes
   ```

3. **Firewall kurallarını kontrol edin:**
   ```bash
   # 6443 portunun açık olduğundan emin olun
   sudo ufw status
   # veya
   sudo iptables -L -n | grep 6443
   ```

**Ingress Nedir ve Ne İşe Yarar?**

Ingress, Kubernetes'te **dış dünyadan cluster içindeki servislere HTTP/HTTPS trafiğini yönlendiren** bir kaynaktır. Bir tür "akıllı reverse proxy" görevi görür.

### Ingress'in Avantajları:

1. **Tek IP Adresi ile Birden Fazla Servis**: 
   - Her servis için ayrı LoadBalancer veya NodePort kullanmak yerine, tek bir IP adresi üzerinden tüm servislere erişebilirsiniz
   - Örnek: `api.barbertime.com`, `admin.barbertime.com` aynı IP'den farklı servislere yönlendirilir

2. **Domain-Based Routing (Alan Adı Tabanlı Yönlendirme)**:
   - Farklı domain'ler farklı servislere yönlendirilebilir
   - Örnek: `api.barbertime.com` → API servisi, `www.barbertime.com` → Web servisi

3. **Path-Based Routing (Yol Tabanlı Yönlendirme)**:
   - Aynı domain altında farklı path'ler farklı servislere yönlendirilebilir
   - Örnek: `barbertime.com/api` → API, `barbertime.com/admin` → Admin paneli

4. **SSL/TLS Terminasyonu**:
   - HTTPS sertifikalarını Ingress seviyesinde yönetebilirsiniz
   - Let's Encrypt gibi otomatik sertifika yönetimi (cert-manager ile)

5. **Load Balancing**:
   - Trafiği otomatik olarak birden fazla pod'a dağıtır
   - Health check'ler yaparak sağlıksız pod'lara trafik göndermez

6. **Maliyet Tasarrufu**:
   - Her servis için ayrı LoadBalancer yerine tek bir LoadBalancer kullanılır
   - Cloud provider'larda LoadBalancer maliyetli olabilir

### Ingress vs Diğer Yöntemler:

| Yöntem | Kullanım | Avantaj | Dezavantaj |
|--------|----------|---------|------------|
| **NodePort** | Test/Development | Basit, ek bileşen gerekmez | Her node'da port açılır, güvenlik riski |
| **LoadBalancer** | Production (tek servis) | Otomatik yük dengeleme | Her servis için ayrı LB, maliyetli |
| **Ingress** | Production (çoklu servis) | Tek IP, domain routing, SSL | Ingress Controller kurulumu gerekir |

### Ingress Nasıl Çalışır?

```
İnternet → Ingress Controller (NGINX) → Service → Pod'lar
                ↓
         Ingress Resource (YAML)
         - Domain: api.barbertime.com
         - Path: /
         - Backend: barbertime-api:80
```

1. **Ingress Controller**: NGINX, Traefik, HAProxy gibi bir reverse proxy
2. **Ingress Resource**: YAML dosyası ile tanımlanan routing kuralları
3. **Service**: Ingress, Service'lere yönlendirme yapar
4. **Pod'lar**: Service, trafiği pod'lara dağıtır

### Örnek Kullanım Senaryosu:

BarberTime uygulamanızda:
- `api.barbertime.local` → API servisine yönlendirilir
- `admin.barbertime.local` → Admin paneli (gelecekte eklenebilir)
- Tek bir Ingress Controller tüm trafiği yönetir

**Not:** Ingress opsiyoneldir. Eğer sadece cluster içinden erişim yeterliyse veya port-forward kullanacaksanız, Ingress kurmanıza gerek yoktur. Ancak production ortamlarında kesinlikle önerilir.

## 4. Docker Image'ları

### 4.1. Docker'a İhtiyaç Var mı?

**Kısa Cevap:** Hayır, eğer image'leri Docker Hub veya başka bir registry'den çekecekseniz Docker kurmanıza gerek yoktur.

**Detaylı Açıklama:**

- **Kubernetes containerd kullanır**: Kubernetes, container runtime olarak containerd kullanır (1.2. bölümünde kurduk)
- **Containerd image'leri çekebilir**: containerd, Docker Hub ve diğer registry'lerden image'leri çekebilir
- **Docker sadece build için gerekli**: Docker'a sadece kendi image'lerinizi build etmek için ihtiyaç vardır

### Senaryo 1: Image'leri Registry'den Çekmek (Önerilen)

Eğer image'leriniz Docker Hub veya başka bir registry'de hazırsa:

✅ **Docker kurmanıza gerek yok!**

Kubernetes deployment dosyalarınızda image adresini belirtmeniz yeterli:

```yaml
# Örnek: k8s/applications/api-deployment.yaml
spec:
  containers:
  - name: api
    image: docker.io/barbertime/barbertime-api:latest  # Docker Hub'dan çeker
    # veya
    image: registry.example.com/barbertime-api:latest  # Private registry'den çeker
```

Kubernetes otomatik olarak image'leri containerd ile çekecektir.

### Senaryo 2: Kendi Image'lerinizi Build Etmek

Eğer kendi image'lerinizi build etmek istiyorsanız:

```bash
# Docker kur
sudo apt install -y docker.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
# Yeni grup yetkilerini almak için logout/login yapın veya:
newgrp docker

# Proje dizininde image'ları build et
docker build -t barbertime-api:latest -f Dockerfile .
docker build -t barbertime-consumer:latest -f Dockerfile.consumer .

# Image'leri Docker Hub'a push et (veya başka bir registry'ye)
docker tag barbertime-api:latest docker.io/kullaniciadi/barbertime-api:latest
docker push docker.io/kullaniciadi/barbertime-api:latest
```

### Senaryo 3: Local Image Kullanmak (Geliştirme için)

Eğer sadece local'de test ediyorsanız ve image'leri build ettiyseniz:

1. Image'leri build edin (Docker gerekli)
2. Kubernetes deployment dosyalarında `imagePullPolicy: IfNotPresent` kullanın
3. Image'leri node'lara manuel yükleyin (opsiyonel):

```bash
# Image'leri containerd'e import et (Docker build sonrası)
sudo ctr -n k8s.io images import barbertime-api.tar
```

**Öneri:** Production ortamında image'leri mutlaka bir registry'de (Docker Hub, GitHub Container Registry, AWS ECR, vb.) tutun. Bu şekilde:
- Tüm node'lar aynı image'i kullanır
- Version kontrolü yapabilirsiniz
- CI/CD pipeline'ları kolaylaşır
- Docker kurmanıza gerek kalmaz

## 5. Kubernetes Yapılandırmalarını Uygulama

### 5.1. Namespace'leri Oluştur

```bash
kubectl apply -f k8s/namespaces/namespaces.yaml
```

### 5.2. Infrastructure Servislerini Kur

```bash
# PostgreSQL
kubectl apply -f k8s/infrastructure/postgresql-secret.yaml
kubectl apply -f k8s/infrastructure/postgresql-configmap.yaml
kubectl apply -f k8s/infrastructure/postgresql-deployment.yaml

# Redis
kubectl apply -f k8s/infrastructure/redis-secret.yaml
kubectl apply -f k8s/infrastructure/redis-deployment.yaml

# Elasticsearch
kubectl apply -f k8s/infrastructure/elasticsearch-configmap.yaml
kubectl apply -f k8s/infrastructure/elasticsearch-deployment.yaml

# RabbitMQ
kubectl apply -f k8s/infrastructure/rabbitmq-secret.yaml
kubectl apply -f k8s/infrastructure/rabbitmq-deployment.yaml
```

### 5.3. Infrastructure Servislerinin Hazır Olmasını Bekle

```bash
# Tüm pod'ların çalışır durumda olduğunu kontrol et
kubectl get pods -n barbertime-dev-infrastructure

# Servislerin hazır olmasını bekle
kubectl wait --for=condition=ready pod -l app=postgresql -n barbertime-dev-infrastructure --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n barbertime-dev-infrastructure --timeout=300s
kubectl wait --for=condition=ready pod -l app=elasticsearch -n barbertime-dev-infrastructure --timeout=300s
kubectl wait --for=condition=ready pod -l app=rabbitmq -n barbertime-dev-infrastructure --timeout=300s
```

### 5.4. Application Servislerini Kur

```bash
# API
kubectl apply -f k8s/applications/api-configmap.yaml
kubectl apply -f k8s/applications/api-secret.yaml
kubectl apply -f k8s/applications/api-deployment.yaml

# Consumer
kubectl apply -f k8s/applications/consumer-configmap.yaml
kubectl apply -f k8s/applications/consumer-secret.yaml
kubectl apply -f k8s/applications/consumer-deployment.yaml

# Ingress (opsiyonel)
kubectl apply -f k8s/applications/ingress.yaml
```

## 6. Kurulumu Doğrulama

### 6.1. Pod Durumlarını Kontrol Et

```bash
# Infrastructure namespace
kubectl get pods -n barbertime-dev-infrastructure

# Application namespace
kubectl get pods -n barbertime-dev
```

### 6.2. Servisleri Kontrol Et

```bash
# Infrastructure servisleri
kubectl get svc -n barbertime-dev-infrastructure

# Application servisleri
kubectl get svc -n barbertime-dev
```

### 6.3. HPA Durumunu Kontrol Et

```bash
kubectl get hpa -n barbertime-dev
```

### 6.4. Logları Kontrol Et

```bash
# API logları
kubectl logs -f deployment/barbertime-api -n barbertime-dev

# Consumer logları
kubectl logs -f deployment/barbertime-consumer -n barbertime-dev
```

## 7. Port Forward ile Test Etme

```bash
# API'ye erişim
kubectl port-forward -n barbertime-dev svc/barbertime-api 8080:80

# Tarayıcıda veya curl ile test et
curl http://localhost:8080/health
```

## 8. Güncelleme ve Bakım

### 8.1. Image Güncelleme

```bash
# Yeni image build et
docker build -t barbertime-api:v1.1.0 -f Dockerfile .

# Deployment'ı güncelle
kubectl set image deployment/barbertime-api api=barbertime-api:v1.1.0 -n barbertime-dev
```

### 8.2. ConfigMap/Secret Güncelleme

```bash
# ConfigMap'i düzenle
kubectl edit configmap api-config -n barbertime-dev

# Pod'ları yeniden başlat (deployment otomatik olarak yeniden başlatır)
kubectl rollout restart deployment/barbertime-api -n barbertime-dev
```

### 8.3. Scaling

HPA otomatik olarak ölçeklendirme yapar, manuel ölçeklendirme için:

```bash
kubectl scale deployment/barbertime-api --replicas=5 -n barbertime-dev
```

## 9. Sorun Giderme

### 9.1. Pod'lar Başlamıyor

```bash
# Pod detaylarını kontrol et
kubectl describe pod <pod-name> -n <namespace>

# Logları kontrol et
kubectl logs <pod-name> -n <namespace>
```

### 9.2. HPA Çalışmıyor

```bash
# Metrics server'ı kontrol et
kubectl get deployment metrics-server -n kube-system
kubectl logs -n kube-system deployment/metrics-server

# HPA detaylarını kontrol et
kubectl describe hpa barbertime-api-hpa -n barbertime-dev
```

### 9.3. Servisler Birbirine Bağlanamıyor

```bash
# DNS çözümlemesini test et
kubectl run -it --rm debug --image=busybox --restart=Never -n barbertime-dev -- nslookup postgresql.barbertime-dev-infrastructure.svc.cluster.local
```

### 9.4. Sistem Pod'ları CrashLoopBackOff Durumunda

Eğer `kube-controller-manager` veya `kube-proxy` CrashLoopBackOff durumundaysa:

**1. Logları kontrol edin:**

```bash
# kube-controller-manager loglarını kontrol et
kubectl logs -n kube-system kube-controller-manager-$(hostname) --tail=50

# kube-proxy loglarını kontrol et
kubectl get pods -n kube-system | grep kube-proxy
kubectl logs -n kube-system kube-proxy-<pod-name> --tail=50
```

**2. kube-controller-manager sorunu:**

Genellikle network veya API server bağlantı sorunlarından kaynaklanır:

```bash
# Pod detaylarını kontrol et
kubectl describe pod -n kube-system kube-controller-manager-$(hostname)

# Eğer "connection refused" hatası varsa, API server'ın çalıştığından emin olun
kubectl get pods -n kube-system | grep kube-apiserver

# API server loglarını kontrol et
kubectl logs -n kube-system kube-apiserver-$(hostname) --tail=50
```

**3. kube-proxy sorunu:**

Genellikle CNI plugin (Flannel) kurulmamışsa veya network yapılandırması eksikse olur:

```bash
# Flannel'in kurulu olduğundan emin olun
kubectl get pods -n kube-flannel

# Eğer Flannel yoksa, kurun
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml

# kube-proxy'yi yeniden başlat
kubectl delete pod -n kube-system -l k8s-app=kube-proxy
```

**4. Metrics Server Pending durumu:**

Metrics server genellikle node'da yeterli kaynak olmadığında veya toleration sorunları nedeniyle Pending kalır:

```bash
# Metrics server'ın neden pending olduğunu kontrol et
kubectl describe pod -n kube-system -l k8s-app=metrics-server

# Eğer node selector veya toleration sorunu varsa, patch uygulayın
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

# Metrics server pod'unu sil ve yeniden oluştur
kubectl delete pod -n kube-system -l k8s-app=metrics-server
```

**5. Tüm sistem pod'larını kontrol et:**

```bash
# Tüm sistem pod'larının durumunu görüntüle
kubectl get pods -n kube-system -o wide

# Her pod için detaylı bilgi
kubectl get pods -n kube-system -o json | jq '.items[] | {name: .metadata.name, status: .status.phase, restarts: .status.containerStatuses[0].restartCount}'
```

**6. API Server Bağlantı Sorunu (Connection Refused):**

Eğer `connection refused` hatası alıyorsanız ve API server çalışmıyorsa:

```bash
# API server pod'unu doğrudan kontrol et (kubectl çalışmıyorsa)
sudo crictl ps -a | grep kube-apiserver

# API server loglarını kontrol et
sudo crictl logs $(sudo crictl ps -a | grep kube-apiserver | awk '{print $1}')

# API server'ın etcd'ye bağlanabildiğini kontrol et
# etcd loglarını kontrol et
sudo crictl logs $(sudo crictl ps -a | grep etcd | awk '{print $1}')

# etcd'nin çalıştığını kontrol et
sudo systemctl status etcd
# veya
sudo crictl ps -a | grep etcd
```

**7. etcd ve API Server Sorunlarını Çözme:**

Eğer etcd CrashLoopBackOff durumundaysa:

```bash
# etcd pod'unun loglarını kontrol et
sudo crictl logs $(sudo crictl ps -a | grep etcd | awk '{print $1}') --tail=100

# etcd veri dizinini kontrol et
ls -la /var/lib/etcd/

# etcd'nin izinlerini kontrol et
sudo ls -la /var/lib/etcd/member/

# Eğer izin sorunu varsa
sudo chown -R etcd:etcd /var/lib/etcd/
# veya
sudo chmod -R 700 /var/lib/etcd/
```

**8. API Server'ı Manuel Başlatma (Acil Durum):**

Eğer API server hiç başlamıyorsa:

```bash
# kubelet'i yeniden başlat (tüm static pod'ları yeniden başlatır)
sudo systemctl restart kubelet

# Birkaç saniye bekleyin
sleep 15

# Pod'ların durumunu kontrol et
sudo crictl ps -a

# API server'ın başladığını kontrol et
sudo crictl ps | grep kube-apiserver
```

**9. Cluster'ı Yeniden Başlatma (Son Çare):**

Eğer tüm yukarıdaki adımlar işe yaramadıysa:

```bash
# Cluster'ı reset et (TÜM VERİLER SİLİNİR!)
sudo kubeadm reset -f

# Temizlik
sudo rm -rf /etc/cni/net.d
sudo rm -rf /var/lib/etcd
sudo rm -rf /var/lib/kubelet
sudo rm -rf ~/.kube

# Yeniden başlat
sudo kubeadm init --pod-network-cidr=10.244.0.0/16

# Kubeconfig'i tekrar kopyala
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# Flannel'i kur
kubectl apply -f https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml
```

## 10. Güvenlik Notları

⚠️ **ÖNEMLİ:** Bu kurulum development ortamı için hazırlanmıştır. Production ortamı için:

1. **Secret'ları güvenli tutun:** Production'da gerçek şifreler kullanın ve secret'ları güvenli bir şekilde yönetin (HashiCorp Vault, AWS Secrets Manager, vb.)
2. **JWT_SECRET'ı değiştirin:** Production'da güçlü bir JWT secret kullanın
3. **Network Policies ekleyin:** Pod'lar arası iletişimi kısıtlayın
4. **RBAC yapılandırın:** Kullanıcı yetkilerini sınırlandırın
5. **TLS/SSL kullanın:** Ingress için TLS sertifikaları ekleyin
6. **Resource Limits:** Production'da daha yüksek limitler ayarlayın
7. **Backup stratejisi:** PostgreSQL ve diğer veritabanları için düzenli backup alın

## 11. Faydalı Komutlar

```bash
# Tüm kaynakları listele
kubectl get all -n barbertime-dev
kubectl get all -n barbertime-dev-infrastructure

# Namespace'i sil (dikkatli kullanın!)
kubectl delete namespace barbertime-dev
kubectl delete namespace barbertime-dev-infrastructure

# Pod'a shell ile bağlan
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh

# Deployment'ı sil ve yeniden oluştur
kubectl delete -f k8s/applications/api-deployment.yaml
kubectl apply -f k8s/applications/api-deployment.yaml
```

## 12. Monitoring (Opsiyonel)

Prometheus ve Grafana kurulumu için:

```bash
# Prometheus Operator kur
kubectl create namespace monitoring
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
```

## Destek

Sorun yaşarsanız, logları kontrol edin ve pod'ların durumunu inceleyin. Kubernetes dokümantasyonunu da referans alabilirsiniz.
