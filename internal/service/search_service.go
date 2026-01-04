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
	Barbers    []models.Barber `json:"barbers,omitempty"`
	Shops      []models.Shop   `json:"shops,omitempty"`
	TotalCount int64           `json:"-"` // Internal use for pagination
}

type searchService struct {
	barberRepo      repository.BarberRepository
	shopRepo        repository.ShopRepository
	shopServiceRepo repository.ShopServiceRepository
}

func NewSearchService(
	barberRepo repository.BarberRepository,
	shopRepo repository.ShopRepository,
	shopServiceRepo repository.ShopServiceRepository,
) SearchService {
	return &searchService{
		barberRepo:      barberRepo,
		shopRepo:        shopRepo,
		shopServiceRepo: shopServiceRepo,
	}
}

func (s *searchService) Search(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, utils.Pagination, error) {
	query = strings.TrimSpace(strings.ToLower(query))

	// Cache key oluştur
	cacheKey := fmt.Sprintf("search:%s:%s:%v:%v:%v:%v:%d:%d", query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)

	// Cache'den kontrol et
	var cachedResults SearchResults
	if err := redis.Get(cacheKey, &cachedResults); err == nil {
		log.Printf("[Search] Cache hit for key: %s", cacheKey)
		pagination := utils.NewPagination(page, limit, cachedResults.TotalCount)
		return cachedResults, pagination, nil
	}

	// Elasticsearch kullanarak arama yap
	if elasticsearch.Client != nil {
		esResults, err := s.searchWithElasticsearch(query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)
		if err == nil {
			log.Printf("[Search] Elasticsearch search successful - barbers: %d, shops: %d, total: %d",
				len(esResults.Barbers), len(esResults.Shops), esResults.TotalCount)
			// Cache'e kaydet (5 dakika)
			if redis.Client != nil {
				redis.Set(cacheKey, esResults, 5*time.Minute)
			}
			pagination := utils.NewPagination(page, limit, esResults.TotalCount)
			return esResults, pagination, nil
		}
		// Elasticsearch hatası varsa fallback'e geç
		log.Printf("[Search] Elasticsearch search failed, falling back to database: %v", err)
	} else {
		log.Printf("[Search] Elasticsearch not available, using database search")
	}

	// Fallback: Database'den arama
	dbResults, pagination, err := s.searchWithDatabase(query, searchType, latitude, longitude, serviceName, maxDistanceKM, page, limit)
	if err != nil {
		log.Printf("[Search] Database search failed: %v", err)
		return SearchResults{}, utils.Pagination{}, err
	}

	log.Printf("[Search] Database search successful - barbers: %d, shops: %d, total: %d",
		len(dbResults.Barbers), len(dbResults.Shops), dbResults.TotalCount)

	// Cache'e kaydet (5 dakika)
	if redis.Client != nil {
		redis.Set(cacheKey, dbResults, 5*time.Minute)
	}

	return dbResults, pagination, nil
}

func (s *searchService) searchWithElasticsearch(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, error) {
	var results SearchResults
	from := (page - 1) * limit

	// Elasticsearch query oluştur
	esQuery := map[string]interface{}{
		"from":  from,
		"size":  limit,
		"query": map[string]interface{}{},
	}

	// Default mesafe 50km
	distanceStr := "50km"
	if maxDistanceKM != nil {
		distanceStr = fmt.Sprintf("%.1fkm", *maxDistanceKM)
	}

	// Bool query için must, filter, should listeleri
	var mustQueries []interface{}
	var filterQueries []interface{}

	// Text search with fuzzy matching
	if query != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": query,
				"fields": []string{
					"name^3",                   // En yüksek öncelik
					"description^2",            // İkinci öncelik
					"bio^2",                    // Barber için
					"services.name^2",          // Servis isimleri (nested)
					"services.description^1.5", // Servis açıklamaları (nested)
					"shopNames^1.5",            // Barber için shop isimleri
					"shopAddresses^1",          // Barber için shop adresleri
					"address^1.5",              // Shop için adres
					"phoneNumber^1",            // Telefon numarası
					"email^1",                  // Email
				},
				"type":                 "best_fields",
				"fuzziness":            "AUTO", // Otomatik fuzzy matching
				"operator":             "or",   // Herhangi bir kelime eşleşirse
				"minimum_should_match": "75%",  // En az %75 kelime eşleşmeli
			},
		})
	}

	// Service filter (nested query with fuzzy search)
	if serviceName != nil && *serviceName != "" {
		serviceNameLower := strings.ToLower(*serviceName)
		serviceFilter := map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "services",
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"should": []interface{}{
							// Exact match (en yüksek öncelik)
							map[string]interface{}{
								"match": map[string]interface{}{
									"services.name": map[string]interface{}{
										"query": serviceNameLower,
										"boost": 3.0,
									},
								},
							},
							// Fuzzy match (benzer isimler)
							map[string]interface{}{
								"match": map[string]interface{}{
									"services.name": map[string]interface{}{
										"query":     serviceNameLower,
										"fuzziness": "AUTO",
										"boost":     1.0,
									},
								},
							},
							// Prefix match (başlangıç eşleşmesi)
							map[string]interface{}{
								"prefix": map[string]interface{}{
									"services.name": map[string]interface{}{
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

	// Sort by distance (eğer location varsa)
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

	var totalCount int64

	// Search barbers
	if searchType == "barber" || searchType == "all" {
		indexName := "barbertime_barbers"

		barberResults, err := elasticsearch.Search(indexName, esQuery, from, limit)
		if err == nil {
			barbers, barberTotal := s.parseBarberResults(barberResults)
			results.Barbers = barbers
			totalCount += barberTotal
		} else {
			log.Printf("[Search] Elasticsearch barber search error: %v", err)
		}
	}

	// Search shops
	if searchType == "shop" || searchType == "all" {
		indexName := "barbertime_shops"

		shopResults, err := elasticsearch.Search(indexName, esQuery, from, limit)
		if err == nil {
			shops, shopTotal := s.parseShopResults(shopResults)
			results.Shops = shops
			totalCount += shopTotal
		} else {
			log.Printf("[Search] Elasticsearch shop search error: %v", err)
		}
	}

	results.TotalCount = totalCount
	return results, nil
}

func (s *searchService) parseBarberResults(esResult map[string]interface{}) ([]models.Barber, int64) {
	var barbers []models.Barber
	var totalCount int64

	hits, ok := esResult["hits"].(map[string]interface{})
	if !ok {
		return barbers, 0
	}

	// Parse total count from Elasticsearch response
	if total, ok := hits["total"].(map[string]interface{}); ok {
		if value, ok := total["value"].(float64); ok {
			totalCount = int64(value)
		}
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return barbers, totalCount
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

		barbers = append(barbers, barber)
	}

	return barbers, totalCount
}

func (s *searchService) parseShopResults(esResult map[string]interface{}) ([]models.Shop, int64) {
	var shops []models.Shop
	var totalCount int64

	hits, ok := esResult["hits"].(map[string]interface{})
	if !ok {
		return shops, 0
	}

	// Parse total count from Elasticsearch response
	if total, ok := hits["total"].(map[string]interface{}); ok {
		if value, ok := total["value"].(float64); ok {
			totalCount = int64(value)
		}
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return shops, totalCount
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

		shops = append(shops, shop)
	}

	return shops, totalCount
}

func (s *searchService) searchWithDatabase(query string, searchType string, latitude, longitude *float64, serviceName *string, maxDistanceKM *float64, page, limit int) (SearchResults, utils.Pagination, error) {
	results := SearchResults{}
	var totalCount int64

	// Konum bilgisi varsa nearby araması yap, yoksa isim bazlı arama yap
	hasLocation := latitude != nil && longitude != nil

	if searchType == "barber" || searchType == "all" {
		var barbers []models.Barber
		var barberTotal int64
		var err error

		if hasLocation {
			barbers, barberTotal, err = s.barberRepo.GetNearby(*latitude, *longitude, maxDistanceKM, page, limit)
			if err != nil {
				log.Printf("[Search] Database barber nearby search error: %v", err)
			}
		} else if query != "" {
			// Konum yoksa isim bazlı arama yap
			barbers, barberTotal, err = s.barberRepo.SearchByName(query, page, limit)
			if err != nil {
				log.Printf("[Search] Database barber name search error: %v", err)
			}
		}

		// İsim ve servis filtreleme (nearby aramasında)
		if hasLocation && (query != "" || (serviceName != nil && *serviceName != "")) {
			var filteredBarbers []models.Barber
			for _, barber := range barbers {
				// Text query filtresi
				if query != "" && !strings.Contains(strings.ToLower(barber.Name), query) {
					continue
				}
				// Service filtresi
				if serviceName != nil && *serviceName != "" {
					if !s.barberHasService(barber, *serviceName) {
						continue
					}
				}
				filteredBarbers = append(filteredBarbers, barber)
			}
			results.Barbers = filteredBarbers
			// Filtreleme sonrası gerçek total bilinmiyor, filtrelenmiş sayıyı kullan
			barberTotal = int64(len(filteredBarbers))
		} else {
			results.Barbers = barbers
		}
		totalCount += barberTotal
	}

	if searchType == "shop" || searchType == "all" {
		var shops []models.Shop
		var shopTotal int64
		var err error

		if hasLocation {
			shops, shopTotal, err = s.shopRepo.GetNearby(*latitude, *longitude, maxDistanceKM, page, limit)
			if err != nil {
				log.Printf("[Search] Database shop nearby search error: %v", err)
			}
		} else if query != "" {
			// Konum yoksa isim bazlı arama yap
			shops, shopTotal, err = s.shopRepo.SearchByName(query, page, limit)
			if err != nil {
				log.Printf("[Search] Database shop name search error: %v", err)
			}
		}

		// İsim ve servis filtreleme (nearby aramasında)
		if hasLocation && (query != "" || (serviceName != nil && *serviceName != "")) {
			var filteredShops []models.Shop
			for _, shop := range shops {
				// Text query filtresi
				if query != "" && !strings.Contains(strings.ToLower(shop.Name), query) {
					continue
				}
				// Service filtresi
				if serviceName != nil && *serviceName != "" {
					if !s.shopHasService(shop.ID, *serviceName) {
						continue
					}
				}
				filteredShops = append(filteredShops, shop)
			}
			results.Shops = filteredShops
			// Filtreleme sonrası gerçek total bilinmiyor, filtrelenmiş sayıyı kullan
			shopTotal = int64(len(filteredShops))
		} else {
			results.Shops = shops
		}
		totalCount += shopTotal
	}

	// Eğer ne konum ne de query yoksa uyarı ver
	if !hasLocation && query == "" {
		log.Printf("[Search] Warning: Database search without location or query - returning empty results")
	}

	results.TotalCount = totalCount
	pagination := utils.NewPagination(page, limit, totalCount)

	return results, pagination, nil
}

// barberHasService checks if a barber has a specific service
func (s *searchService) barberHasService(barber models.Barber, serviceName string) bool {
	serviceNameLower := strings.ToLower(serviceName)
	for _, shop := range barber.Shops {
		shopServices, err := s.shopServiceRepo.GetByShopID(shop.ID)
		if err == nil {
			for _, shopService := range shopServices {
				shopServiceNameLower := strings.ToLower(shopService.Service.Name)
				if strings.Contains(shopServiceNameLower, serviceNameLower) ||
					strings.Contains(serviceNameLower, shopServiceNameLower) {
					return true
				}
			}
		}
	}
	return false
}

// shopHasService checks if a shop has a specific service
func (s *searchService) shopHasService(shopID uint, serviceName string) bool {
	serviceNameLower := strings.ToLower(serviceName)
	shopServices, err := s.shopServiceRepo.GetByShopID(shopID)
	if err != nil {
		return false
	}
	for _, shopService := range shopServices {
		shopServiceNameLower := strings.ToLower(shopService.Service.Name)
		if strings.Contains(shopServiceNameLower, serviceNameLower) ||
			strings.Contains(serviceNameLower, shopServiceNameLower) {
			return true
		}
	}
	return false
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

	// Services listesi (nested objects)
	var services []map[string]interface{}
	for _, shopService := range shopServices {
		serviceDoc := map[string]interface{}{
			"id":       shopService.ServiceID,
			"name":     shopService.Service.Name,
			"price":    shopService.Price,
			"duration": shopService.Duration,
		}
		if shopService.Description != nil {
			serviceDoc["description"] = *shopService.Description
		}
		services = append(services, serviceDoc)
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
		"services": services,
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
