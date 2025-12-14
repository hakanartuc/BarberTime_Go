#!/bin/bash

# Kubernetes Repository Düzeltme Scripti
# Eski repository'yi kaldırıp yeni repository'yi ekler

set -e

echo "Eski Kubernetes repository'si kaldırılıyor..."
sudo rm -f /etc/apt/sources.list.d/kubernetes.list
sudo rm -f /usr/share/keyrings/kubernetes-archive-keyring.gpg

echo "Gerekli paketler kuruluyor..."
sudo apt install -y apt-transport-https ca-certificates curl gpg

echo "Yeni Kubernetes repository ekleniyor..."
# GPG key'i ekle
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# Repository ekle
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

echo "Repository güncelleniyor..."
sudo apt update

echo "Kubernetes bileşenleri kuruluyor..."
sudo apt install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

echo "✓ Kubernetes repository düzeltildi ve kurulum tamamlandı!"
