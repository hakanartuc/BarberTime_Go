#!/bin/bash

# BarberTime Kubernetes Kurulum Scripti
# Bu script tüm Kubernetes yapılandırmalarını sırayla uygular

set -e

echo "========================================="
echo "BarberTime Kubernetes Kurulumu Başlıyor"
echo "========================================="

# Renkli çıktı için
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Namespace'leri oluştur
echo -e "${YELLOW}[1/4] Namespace'ler oluşturuluyor...${NC}"
kubectl apply -f namespaces/namespaces.yaml
echo -e "${GREEN}✓ Namespace'ler oluşturuldu${NC}"

# 2. Infrastructure servislerini kur
echo -e "${YELLOW}[2/4] Infrastructure servisleri kuruluyor...${NC}"

echo "  - PostgreSQL kuruluyor..."
kubectl apply -f infrastructure/postgresql-secret.yaml
kubectl apply -f infrastructure/postgresql-configmap.yaml
kubectl apply -f infrastructure/postgresql-deployment.yaml

echo "  - Redis kuruluyor..."
kubectl apply -f infrastructure/redis-secret.yaml
kubectl apply -f infrastructure/redis-deployment.yaml

echo "  - Elasticsearch kuruluyor..."
kubectl apply -f infrastructure/elasticsearch-configmap.yaml
kubectl apply -f infrastructure/elasticsearch-deployment.yaml

echo "  - RabbitMQ kuruluyor..."
kubectl apply -f infrastructure/rabbitmq-secret.yaml
kubectl apply -f infrastructure/rabbitmq-deployment.yaml

echo -e "${GREEN}✓ Infrastructure servisleri kuruldu${NC}"

# 3. Infrastructure servislerinin hazır olmasını bekle
echo -e "${YELLOW}[3/4] Infrastructure servislerinin hazır olması bekleniyor...${NC}"
echo "  Bu işlem birkaç dakika sürebilir..."

kubectl wait --for=condition=ready pod -l app=postgresql \
  -n barbertime-dev-infrastructure --timeout=300s || true

kubectl wait --for=condition=ready pod -l app=redis \
  -n barbertime-dev-infrastructure --timeout=300s || true

kubectl wait --for=condition=ready pod -l app=elasticsearch \
  -n barbertime-dev-infrastructure --timeout=300s || true

kubectl wait --for=condition=ready pod -l app=rabbitmq \
  -n barbertime-dev-infrastructure --timeout=300s || true

echo -e "${GREEN}✓ Infrastructure servisleri hazır${NC}"

# 4. Application servislerini kur
echo -e "${YELLOW}[4/4] Application servisleri kuruluyor...${NC}"

echo "  - API kuruluyor..."
kubectl apply -f applications/api-configmap.yaml
kubectl apply -f applications/api-secret.yaml
kubectl apply -f applications/api-deployment.yaml

echo "  - Consumer kuruluyor..."
kubectl apply -f applications/consumer-configmap.yaml
kubectl apply -f applications/consumer-secret.yaml
kubectl apply -f applications/consumer-deployment.yaml

echo "  - Ingress kuruluyor (opsiyonel)..."
kubectl apply -f applications/ingress.yaml || echo "  Ingress zaten mevcut veya kurulamadı"

echo -e "${GREEN}✓ Application servisleri kuruldu${NC}"

echo ""
echo "========================================="
echo -e "${GREEN}Kurulum Tamamlandı!${NC}"
echo "========================================="
echo ""
echo "Durum kontrolü için:"
echo "  kubectl get pods -n barbertime-dev-infrastructure"
echo "  kubectl get pods -n barbertime-dev"
echo "  kubectl get hpa -n barbertime-dev"
echo ""
echo "Logları görüntülemek için:"
echo "  kubectl logs -f deployment/barbertime-api -n barbertime-dev"
echo "  kubectl logs -f deployment/barbertime-consumer -n barbertime-dev"
echo ""
