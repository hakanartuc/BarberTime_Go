package service

import (
	"barbertime_go/internal/models"
	"barbertime_go/internal/repository"
	"barbertime_go/internal/utils"
	"barbertime_go/pkg/elasticsearch"
	"barbertime_go/pkg/redis"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type SearchService interface {
	Search(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, utils.Pagination, error)
	IndexBarber(barber *models.Barber) error
	IndexShop(shop *models.Shop) error
	DeleteBarberFromIndex(barberID uint) error
	DeleteShopFromIndex(shopID uint) error
}

type SearchResults struct {
	Barbers []models.Barber `json:"barbers,omitempty"`
	Shops   []models.Shop   `json:"shops,omitempty"`
}

type searchService struct {
	barberRepo      repository.BarberRepository
	shopRepo        repository.ShopRepository
	shopServiceRepo repository.ShopServiceRepository
	barberShopRepo  repository.BarberShopRepository
}

func NewSearchService(
	barberRepo repository.BarberRepository,
	shopRepo repository.ShopRepository,
	shopServiceRepo repository.ShopServiceRepository,
	barberShopRepo repository.BarberShopRepository,
) SearchService {
	return &searchService{
		barberRepo:      barberRepo,
		shopRepo:        shopRepo,
		shopServiceRepo: shopServiceRepo,
		barberShopRepo:  barberShopRepo,
	}
}

func (s *searchService) Search(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, utils.Pagination, error) {
	query = strings.TrimSpace(strings.ToLower(query))

	// Cache key oluştur
	cacheKey := fmt.Sprintf("search:%s:%s:%v:%v:%v:%v:%d:%d", query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)

	// Cache'den kontrol et
	var cachedResults SearchResults
	if err := redis.Get(cacheKey, &cachedResults); err == nil {
		// Cache'den bulundu, pagination hesapla
		total := int64(len(cachedResults.Barbers) + len(cachedResults.Shops))
		pagination := utils.NewPagination(page, limit, total)
		return cachedResults, pagination, nil
	}

	// Elasticsearch kullanarak arama yap
	if elasticsearch.Client != nil {
		esResults, err := s.searchWithElasticsearch(query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)
		if err == nil {
			// Cache'e kaydet (5 dakika)
			if redis.Client != nil {
				redis.Set(cacheKey, esResults, 5*time.Minute)
			}
			total := int64(len(esResults.Barbers) + len(esResults.Shops))
			pagination := utils.NewPagination(page, limit, total)
			return esResults, pagination, nil
		}
		// Elasticsearch hatası varsa fallback'e geç
	}

	// Fallback: Database'den arama
	return s.searchWithDatabase(query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)
}

func (s *searchService) searchWithElasticsearch(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, error) {
	var results SearchResults
	from := (page - 1) * limit

	// Default mesafe 50km
	distanceStr := "50km"
	if maxDistanceKM != nil {
		distanceStr = fmt.Sprintf("%.1fkm", *maxDistanceKM)
	}

	// Barber için query oluştur
	buildQuery := func(fields []string) map[string]interface{} {
		esQuery := map[string]interface{}{
			"from":  from,
			"size":  limit,
			"query": map[string]interface{}{},
		}

		var mustQueries []interface{}
		var filterQueries []interface{}

		// Text search with fuzzy matching
		if query != "" {
			mustQueries = append(mustQueries, map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":                query,
					"fields":               fields,
					"type":                 "best_fields",
					"fuzziness":            "AUTO",
					"operator":             "or",
					"minimum_should_match": "75%",
				},
			})
		}

		// Service filter (nested query with fuzzy search)
		if serviceName != nil && *serviceName != "" {
			serviceNameLower := strings.ToLower(*serviceName)
			serviceFilter := map[string]interface{}{
				"nested": map[string]interface{}{
					"path": "shopServices",
					"query": map[string]interface{}{
						"bool": map[string]interface{}{
							"should": []interface{}{
								map[string]interface{}{
									"match": map[string]interface{}{
										"shopServices.serviceName": map[string]interface{}{
											"query": serviceNameLower,
											"boost": 3.0,
										},
									},
								},
								map[string]interface{}{
									"match": map[string]interface{}{
										"shopServices.serviceName": map[string]interface{}{
											"query":     serviceNameLower,
											"fuzziness": "AUTO",
											"boost":     1.0,
										},
									},
								},
								map[string]interface{}{
									"prefix": map[string]interface{}{
										"shopServices.serviceName": map[string]interface{}{
											"value": serviceNameLower,
											"boost": 2.0,
										},
									},
								},
							},
							"minimum_should_match": 1,
						},
					},
				},
			}
			mustQueries = append(mustQueries, serviceFilter)
		}

		// Location-based search
		if latitude != nil && longitude != nil {
			geoQuery := map[string]interface{}{
				"geo_distance": map[string]interface{}{
					"distance": distanceStr,
					"location": map[string]float64{
						"lat": *latitude,
						"lon": *longitude,
					},
				},
			}
			filterQueries = append(filterQueries, geoQuery)
		}

		// Query'yi birleştir
		if len(mustQueries) > 0 || len(filterQueries) > 0 {
			boolQuery := map[string]interface{}{}
			if len(mustQueries) > 0 {
				boolQuery["must"] = mustQueries
			}
			if len(filterQueries) > 0 {
				boolQuery["filter"] = filterQueries
			}
			esQuery["query"] = map[string]interface{}{
				"bool": boolQuery,
			}
		} else {
			esQuery["query"] = map[string]interface{}{
				"match_all": map[string]interface{}{},
			}
		}

		// Sort by distance
		if latitude != nil && longitude != nil {
			esQuery["sort"] = []interface{}{
				map[string]interface{}{
					"_geo_distance": map[string]interface{}{
						"location": map[string]float64{
							"lat": *latitude,
							"lon": *longitude,
						},
						"order":         "asc",
						"unit":          "km",
						"distance_type": "plane",
					},
				},
			}
		}

		return esQuery
	}

	// Barber alanları
	barberFields := []string{
		"name^3",
		"username^3",
		"bio^2",
		"shopNames^1.5",
		"shopAddresses^1",
	}

	// Shop alanları
	shopFields := []string{
		"name^3",
		"description^2",
		"address^1.5",
		"phoneNumber^1",
		"email^1",
	}

	// Search barbers
	if searchType == "barber" || searchType == "all" {
		barberQuery := buildQuery(barberFields)
		barberResults, err := elasticsearch.Search("barbertime_barbers", barberQuery, from, limit)
		if err != nil {
			log.Printf("Elasticsearch barber search error: %v", err)
		} else {
			barbers := s.parseBarberResults(barberResults)
			results.Barbers = barbers
			log.Printf("Barber search returned %d results", len(barbers))
		}
	}

	// Search shops
	if searchType == "shop" || searchType == "all" {
		shopQuery := buildQuery(shopFields)
		shopResults, err := elasticsearch.Search("barbertime_shops", shopQuery, from, limit)
		if err != nil {
			log.Printf("Elasticsearch shop search error: %v", err)
		} else {
			shops := s.parseShopResults(shopResults)
			results.Shops = shops
			log.Printf("Shop search returned %d results", len(shops))
		}
	}

	return results, nil
}

func (s *searchService) parseBarberResults(esResult map[string]interface{}) []models.Barber {
	var barbers []models.Barber

	hits, ok := esResult["hits"].(map[string]interface{})
	if !ok {
		return barbers
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return barbers
	}

	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		// Elasticsearch'ten gelen veriyi direkt model'e map et
		barber := models.Barber{}

		// ID
		if idFloat, ok := source["id"].(float64); ok {
			barber.ID = uint(idFloat)
		}

		// UserID
		if userIdFloat, ok := source["userId"].(float64); ok {
			barber.UserID = uint(userIdFloat)
		}

		// Username
		if username, ok := source["username"].(string); ok {
			barber.Username = username
		}

		// Name
		if name, ok := source["name"].(string); ok {
			barber.Name = name
		}

		// Bio
		if bio, ok := source["bio"].(string); ok {
			barber.Bio = bio
		}

		// Experience
		if exp, ok := source["experience"].(float64); ok {
			barber.Experience = int(exp)
		}

		// Rating
		if ratingMap, ok := source["rating"].(map[string]interface{}); ok {
			if avg, ok := ratingMap["average"].(float64); ok {
				barber.Rating.Average = avg
			}
			if total, ok := ratingMap["totalReview"].(float64); ok {
				barber.Rating.TotalReview = int(total)
			}
		}

		// Avatar
		if avatar, ok := source["avatar"].(string); ok {
			barber.Avatar = &avatar
		}

		// Services bilgisi artık shop'lardan gelecek, barber'ın direkt servisleri yok

		barbers = append(barbers, barber)
	}

	return barbers
}

func (s *searchService) parseShopResults(esResult map[string]interface{}) []models.Shop {
	var shops []models.Shop

	hits, ok := esResult["hits"].(map[string]interface{})
	if !ok {
		return shops
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return shops
	}

	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		// Elasticsearch'ten gelen veriyi direkt model'e map et
		shop := models.Shop{}

		// ID
		if idFloat, ok := source["id"].(float64); ok {
			shop.ID = uint(idFloat)
		}

		// Name
		if name, ok := source["name"].(string); ok {
			shop.Name = name
		}

		// Description
		if desc, ok := source["description"].(string); ok {
			shop.Description = desc
		}

		// PhoneNumber
		if phone, ok := source["phoneNumber"].(string); ok {
			shop.PhoneNumber = phone
		}

		// Email
		if email, ok := source["email"].(string); ok {
			shop.Email = email
		}

		// Website
		if website, ok := source["website"].(string); ok {
			shop.Website = website
		}

		// WorkingHours
		if wh, ok := source["workingHours"].(string); ok {
			shop.WorkingHours = wh
		}

		// ClosedDays
		if cd, ok := source["closedDays"].(string); ok {
			shop.ClosedDays = cd
		}

		// Location
		if locMap, ok := source["location"].(map[string]interface{}); ok {
			if lat, ok := locMap["lat"].(float64); ok {
				shop.Location.Latitude = lat
			}
			if lon, ok := locMap["lon"].(float64); ok {
				shop.Location.Longitude = lon
			}
		}

		// Address
		if addr, ok := source["address"].(string); ok {
			shop.Location.Address = addr
		}

		// Status (confirmed olarak işaretle, çünkü sadece confirmed shop'lar indexleniyor)
		shop.Status = models.ShopStatusConfirmed

		// Owners bilgilerini Elasticsearch'ten parse et
		if ownersList, ok := source["owners"].([]interface{}); ok {
			var owners []models.User
			for _, ownerItem := range ownersList {
				if ownerMap, ok := ownerItem.(map[string]interface{}); ok {
					owner := models.User{}
					if idFloat, ok := ownerMap["id"].(float64); ok {
						owner.ID = uint(idFloat)
					}
					if name, ok := ownerMap["name"].(string); ok {
						owner.Name = name
					}
					if avatar, ok := ownerMap["avatar"].(string); ok {
						owner.Avatar = &avatar
					}
					owners = append(owners, owner)
				}
			}
			shop.Owners = owners
		}

		// ShopServices bilgilerini Elasticsearch'ten parse et
		if servicesList, ok := source["shopServices"].([]interface{}); ok {
			var shopServices []models.ShopService
			for _, serviceItem := range servicesList {
				if serviceMap, ok := serviceItem.(map[string]interface{}); ok {
					shopService := models.ShopService{}
					if idFloat, ok := serviceMap["id"].(float64); ok {
						shopService.ID = uint(idFloat)
					}
					if serviceIdFloat, ok := serviceMap["serviceId"].(float64); ok {
						shopService.ServiceID = uint(serviceIdFloat)
					}
					if serviceName, ok := serviceMap["serviceName"].(string); ok {
						shopService.Service.Name = serviceName
					}
					if serviceImage, ok := serviceMap["serviceImage"].(string); ok {
						shopService.Service.Image = &serviceImage
					}
					if desc, ok := serviceMap["description"].(string); ok {
						shopService.Description = &desc
					}
					if price, ok := serviceMap["price"].(float64); ok {
						shopService.Price = price
					}
					if duration, ok := serviceMap["duration"].(float64); ok {
						shopService.Duration = int(duration)
					}
					shopServices = append(shopServices, shopService)
				}
			}
			shop.ShopServices = shopServices
		}

		shops = append(shops, shop)
	}

	return shops
}

func (s *searchService) searchWithDatabase(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, utils.Pagination, error) {
	results := SearchResults{}

	// Eğer konum bilgisi varsa, nearby araması yap
	if latitude != nil && longitude != nil {
		if searchType == "barber" || searchType == "all" {
			barbers, _, _ := s.barberRepo.GetNearby(*latitude, *longitude, maxDistanceKM, page, limit)
			// İsim bazlı filtreleme (basit contains - Elasticsearch kullanılamazsa)
			var filteredBarbers []models.Barber
			for _, barber := range barbers {
				// Text query filtresi (basit contains) - name veya username ile eşleşme
				if query != "" && !strings.Contains(strings.ToLower(barber.Name), query) && !strings.Contains(strings.ToLower(barber.Username), query) {
					continue
				}
				// Service filtresi (basit contains - Elasticsearch kullanılamazsa)
				if serviceName != nil && *serviceName != "" {
					serviceNameLower := strings.ToLower(*serviceName)
					hasService := false
					for _, shop := range barber.Shops {
						shopServices, err := s.shopServiceRepo.GetByShopID(shop.ID)
						if err == nil {
							for _, shopService := range shopServices {
								// Basit contains kontrolü (Elasticsearch olmadığında)
								shopServiceNameLower := strings.ToLower(shopService.Service.Name)
								if strings.Contains(shopServiceNameLower, serviceNameLower) ||
									strings.Contains(serviceNameLower, shopServiceNameLower) {
									hasService = true
									break
								}
							}
						}
						if hasService {
							break
						}
					}
					if !hasService {
						continue
					}
				}
				filteredBarbers = append(filteredBarbers, barber)
			}
			results.Barbers = filteredBarbers
		}

		if searchType == "shop" || searchType == "all" {
			shops, _, _ := s.shopRepo.GetNearby(*latitude, *longitude, maxDistanceKM, page, limit)
			// İsim bazlı filtreleme (basit contains - Elasticsearch kullanılamazsa)
			var filteredShops []models.Shop
			for _, shop := range shops {
				// Text query filtresi (basit contains)
				if query != "" && !strings.Contains(strings.ToLower(shop.Name), query) {
					continue
				}
				// Service filtresi (basit contains - Elasticsearch kullanılamazsa)
				if serviceName != nil && *serviceName != "" {
					serviceNameLower := strings.ToLower(*serviceName)
					shopServices, err := s.shopServiceRepo.GetByShopID(shop.ID)
					if err != nil {
						continue
					}
					hasService := false
					for _, shopService := range shopServices {
						// Basit contains kontrolü (Elasticsearch olmadığında)
						shopServiceNameLower := strings.ToLower(shopService.Service.Name)
						if strings.Contains(shopServiceNameLower, serviceNameLower) ||
							strings.Contains(serviceNameLower, shopServiceNameLower) {
							hasService = true
							break
						}
					}
					if !hasService {
						continue
					}
				}
				filteredShops = append(filteredShops, shop)
			}
			results.Shops = filteredShops
		}
	} else {
		// Konum bilgisi yoksa, Elasticsearch kullanılması önerilir
		// Database fallback bu durumda sınırlıdır
		// Sadece boş sonuç döndürüyoruz (Elasticsearch olmadığında konum bilgisi şart)
		log.Printf("Warning: Database fallback without location - Elasticsearch recommended")
	}

	// Toplam sayıyı hesapla
	total := int64(len(results.Barbers) + len(results.Shops))
	pagination := utils.NewPagination(page, limit, total)

	return results, pagination, nil
}

// fuzzyMatch basit bir fuzzy matching algoritması (Levenshtein distance benzeri)
// İki string arasında benzerlik kontrolü yapar
func (s *searchService) fuzzyMatch(str1, str2 string) bool {
	// Eğer biri diğerinin substring'i ise true döndür
	if strings.Contains(str1, str2) || strings.Contains(str2, str1) {
		return true
	}

	// Basit karakter benzerliği kontrolü (en az %70 benzerlik)
	len1 := len(str1)
	len2 := len(str2)
	if len1 == 0 || len2 == 0 {
		return false
	}

	// Kısa string'i uzun string ile karşılaştır
	shorter := str1
	longer := str2
	if len1 > len2 {
		shorter = str2
		longer = str1
	}

	// Ortak karakter sayısını hesapla
	commonChars := 0
	for _, char := range shorter {
		if strings.ContainsRune(longer, char) {
			commonChars++
		}
	}

	// Benzerlik oranı hesapla
	similarity := float64(commonChars) / float64(len(shorter))
	return similarity >= 0.7
}

func (s *searchService) IndexBarber(barber *models.Barber) error {
	if elasticsearch.Client == nil {
		return fmt.Errorf("elasticsearch client is not available")
	}

	indexName := "barbertime_barbers"

	// Index mapping oluştur (ilk seferinde)
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"userId": map[string]interface{}{
					"type": "long",
				},
				"username": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"bio": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"experience": map[string]interface{}{
					"type": "integer",
				},
				"rating": map[string]interface{}{
					"properties": map[string]interface{}{
						"average": map[string]interface{}{
							"type": "float",
						},
						"totalReview": map[string]interface{}{
							"type": "integer",
						},
					},
				},
				"avatar": map[string]interface{}{
					"type": "keyword",
				},
				"services": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "long",
						},
						"name": map[string]interface{}{
							"type":     "text",
							"analyzer": "standard",
						},
						"description": map[string]interface{}{
							"type":     "text",
							"analyzer": "standard",
						},
						"price": map[string]interface{}{
							"type": "float",
						},
						"duration": map[string]interface{}{
							"type": "integer",
						},
					},
				},
				"shopNames": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"shopAddresses": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"location": map[string]interface{}{
					"type": "geo_point",
				},
			},
		},
	}
	if err := elasticsearch.CreateIndex(indexName, mapping); err != nil {
		log.Printf("Warning: Failed to create index %s: %v", indexName, err)
	}

	// Document oluştur
	doc := map[string]interface{}{
		"id":         barber.ID,
		"userId":     barber.UserID,
		"username":   barber.Username,
		"name":       barber.Name,
		"bio":        barber.Bio,
		"experience": barber.Experience,
		"rating": map[string]interface{}{
			"average":     barber.Rating.Average,
			"totalReview": barber.Rating.TotalReview,
		},
	}

	if barber.Avatar != nil {
		doc["avatar"] = *barber.Avatar
	}

	// Barber'ın shop'larından servisleri topla
	var allServices []map[string]interface{}
	serviceMap := make(map[uint]map[string]interface{}) // serviceID -> serviceDoc (duplicate'ları önlemek için)

	for _, shop := range barber.Shops {
		shopServices, err := s.shopServiceRepo.GetByShopID(shop.ID)
		if err == nil {
			for _, shopService := range shopServices {
				// Eğer bu service daha önce eklenmemişse ekle
				if _, exists := serviceMap[shopService.ServiceID]; !exists {
					serviceDoc := map[string]interface{}{
						"id":       shopService.ServiceID,
						"name":     shopService.Service.Name,
						"price":    shopService.Price,
						"duration": shopService.Duration,
					}
					if shopService.Description != nil {
						serviceDoc["description"] = *shopService.Description
					}
					serviceMap[shopService.ServiceID] = serviceDoc
					allServices = append(allServices, serviceDoc)
				}
			}
		}
	}
	doc["services"] = allServices

	// Shop bilgileri
	var shopNames []string
	var shopAddresses []string
	var firstShopLocation *models.Location

	if len(barber.Shops) > 0 {
		for _, shop := range barber.Shops {
			shopNames = append(shopNames, shop.Name)
			if shop.Location.Address != "" {
				shopAddresses = append(shopAddresses, shop.Location.Address)
			}
			// İlk shop'un location'ını kullan
			if firstShopLocation == nil && shop.Location.Latitude != 0 && shop.Location.Longitude != 0 {
				firstShopLocation = &shop.Location
			}
		}
		doc["shopNames"] = shopNames
		if len(shopAddresses) > 0 {
			doc["shopAddresses"] = shopAddresses
		}
	}

	// Location (ilk shop'un location'ı)
	if firstShopLocation != nil {
		doc["location"] = map[string]interface{}{
			"lat": firstShopLocation.Latitude,
			"lon": firstShopLocation.Longitude,
		}
		if firstShopLocation.Address != "" {
			doc["address"] = firstShopLocation.Address
		}
	}

	if err := elasticsearch.IndexDocument(indexName, strconv.FormatUint(uint64(barber.ID), 10), doc); err != nil {
		return err
	}

	log.Printf("Successfully indexed barber %d in Elasticsearch index %s", barber.ID, indexName)
	return nil
}

func (s *searchService) IndexShop(shop *models.Shop) error {
	if elasticsearch.Client == nil {
		return fmt.Errorf("elasticsearch client is not available")
	}

	indexName := "barbertime_shops"

	// Index mapping oluştur
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "long",
				},
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"phoneNumber": map[string]interface{}{
					"type": "keyword",
				},
				"email": map[string]interface{}{
					"type": "keyword",
				},
				"website": map[string]interface{}{
					"type": "keyword",
				},
				"workingHours": map[string]interface{}{
					"type": "text",
				},
				"closedDays": map[string]interface{}{
					"type": "text",
				},
				"address": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"location": map[string]interface{}{
					"type": "geo_point",
				},
				"owners": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "long",
						},
						"name": map[string]interface{}{
							"type": "keyword",
						},
						"avatar": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"shopServices": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "long",
						},
						"serviceId": map[string]interface{}{
							"type": "long",
						},
						"serviceName": map[string]interface{}{
							"type":     "text",
							"analyzer": "standard",
						},
						"serviceImage": map[string]interface{}{
							"type": "keyword",
						},
						"description": map[string]interface{}{
							"type":     "text",
							"analyzer": "standard",
						},
						"price": map[string]interface{}{
							"type": "float",
						},
						"duration": map[string]interface{}{
							"type": "integer",
						},
					},
				},
			},
		},
	}
	if err := elasticsearch.CreateIndex(indexName, mapping); err != nil {
		log.Printf("Warning: Failed to create index %s: %v", indexName, err)
	}

	// Shop'un servislerini al
	shopServices, err := s.shopServiceRepo.GetByShopID(shop.ID)
	if err != nil {
		log.Printf("Warning: Failed to get services for shop %d: %v", shop.ID, err)
	}

	// ShopServices listesi (nested objects)
	var shopServicesDoc []map[string]interface{}
	for _, shopService := range shopServices {
		serviceDoc := map[string]interface{}{
			"id":          shopService.ID,
			"serviceId":   shopService.ServiceID,
			"serviceName": shopService.Service.Name,
			"price":       shopService.Price,
			"duration":    shopService.Duration,
		}
		if shopService.Service.Image != nil {
			serviceDoc["serviceImage"] = *shopService.Service.Image
		}
		if shopService.Description != nil {
			serviceDoc["description"] = *shopService.Description
		}
		shopServicesDoc = append(shopServicesDoc, serviceDoc)
	}

	// Owner bilgilerini al
	owners, err := s.barberShopRepo.GetOwners(shop.ID)
	if err != nil {
		log.Printf("Warning: Failed to get owners for shop %d: %v", shop.ID, err)
	}

	// Owners listesi (nested objects)
	var ownersDoc []map[string]interface{}
	for _, owner := range owners {
		ownerDoc := map[string]interface{}{
			"id":   owner.ID,
			"name": owner.Name,
		}
		if owner.Avatar != nil {
			ownerDoc["avatar"] = *owner.Avatar
		}
		ownersDoc = append(ownersDoc, ownerDoc)
	}

	// Document oluştur
	doc := map[string]interface{}{
		"id":          shop.ID,
		"name":        shop.Name,
		"description": shop.Description,
		"phoneNumber": shop.PhoneNumber,
		"email":       shop.Email,
		"website":     shop.Website,
		"location": map[string]float64{
			"lat": shop.Location.Latitude,
			"lon": shop.Location.Longitude,
		},
		"shopServices": shopServicesDoc,
		"owners":       ownersDoc,
	}

	if shop.Location.Address != "" {
		doc["address"] = shop.Location.Address
	}

	if shop.WorkingHours != "" {
		doc["workingHours"] = shop.WorkingHours
	}

	if shop.ClosedDays != "" {
		doc["closedDays"] = shop.ClosedDays
	}

	if err := elasticsearch.IndexDocument(indexName, strconv.FormatUint(uint64(shop.ID), 10), doc); err != nil {
		return err
	}

	log.Printf("Successfully indexed shop %d in Elasticsearch index %s", shop.ID, indexName)
	return nil
}

func (s *searchService) DeleteBarberFromIndex(barberID uint) error {
	if elasticsearch.Client == nil {
		return nil
	}
	indexName := "barbertime_barbers"
	return elasticsearch.DeleteDocument(indexName, strconv.FormatUint(uint64(barberID), 10))
}

func (s *searchService) DeleteShopFromIndex(shopID uint) error {
	if elasticsearch.Client == nil {
		return nil
	}
	indexName := "barbertime_shops"
	return elasticsearch.DeleteDocument(indexName, strconv.FormatUint(uint64(shopID), 10))
}
