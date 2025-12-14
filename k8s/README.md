# Kubernetes Yapılandırma Dosyaları

Bu dizin, BarberTime uygulamasının Kubernetes ortamında çalışması için gerekli tüm yapılandırma dosyalarını içerir.

## Dizin Yapısı

```
k8s/
├── namespaces/
│   └── namespaces.yaml              # Namespace tanımlamaları
├── infrastructure/
│   ├── postgresql-secret.yaml       # PostgreSQL şifreleri
│   ├── postgresql-configmap.yaml   # PostgreSQL yapılandırması
│   ├── postgresql-deployment.yaml   # PostgreSQL StatefulSet ve Service
│   ├── redis-secret.yaml            # Redis şifreleri
│   ├── redis-deployment.yaml        # Redis Deployment ve Service
│   ├── elasticsearch-configmap.yaml # Elasticsearch yapılandırması
│   ├── elasticsearch-deployment.yaml # Elasticsearch StatefulSet ve Service
│   ├── rabbitmq-secret.yaml         # RabbitMQ kullanıcı bilgileri
│   └── rabbitmq-deployment.yaml     # RabbitMQ StatefulSet ve Service
└── applications/
    ├── api-configmap.yaml           # API uygulama yapılandırması
    ├── api-secret.yaml              # API gizli bilgileri
    ├── api-deployment.yaml          # API Deployment, Service ve HPA
    ├── consumer-configmap.yaml      # Consumer yapılandırması
    ├── consumer-secret.yaml         # Consumer gizli bilgileri
    ├── consumer-deployment.yaml     # Consumer Deployment ve HPA
    └── ingress.yaml                 # Ingress yapılandırması (opsiyonel)
```

## Namespace'ler

- **barbertime-dev-infrastructure**: Altyapı servisleri (PostgreSQL, Redis, Elasticsearch, RabbitMQ)
- **barbertime-dev**: Uygulama servisleri (API, Consumer)

## Kurulum Sırası

1. Namespace'leri oluştur
2. Infrastructure servislerini kur (PostgreSQL, Redis, Elasticsearch, RabbitMQ)
3. Infrastructure servislerinin hazır olmasını bekle
4. Application servislerini kur (API, Consumer)

Detaylı kurulum talimatları için ana dizindeki `KUBERNETES_SETUP.md` dosyasına bakın.

## Güvenlik Notları

⚠️ **ÖNEMLİ:** Bu yapılandırmalar development ortamı için hazırlanmıştır. Production ortamı için:

- Secret'lardaki şifreleri güçlü değerlerle değiştirin
- JWT_SECRET'ı güvenli bir şekilde oluşturun
- Network Policies ekleyin
- RBAC yapılandırması yapın
- TLS/SSL sertifikaları kullanın
