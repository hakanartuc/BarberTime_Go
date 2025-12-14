#!/bin/bash

# BarberTime Kubernetes Silme Scripti
# DİKKAT: Bu script tüm kaynakları siler!

set -e

echo "========================================="
echo "BarberTime Kubernetes Kaynakları Siliniyor"
echo "========================================="
echo ""
read -p "Tüm kaynakları silmek istediğinizden emin misiniz? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "İşlem iptal edildi."
    exit 0
fi

# Application namespace'i sil
echo "Application namespace siliniyor..."
kubectl delete namespace barbertime-dev --ignore-not-found=true

# Infrastructure namespace'i sil
echo "Infrastructure namespace siliniyor..."
kubectl delete namespace barbertime-dev-infrastructure --ignore-not-found=true

echo ""
echo "========================================="
echo "Tüm kaynaklar silindi!"
echo "========================================="
