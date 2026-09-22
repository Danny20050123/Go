package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const pokemonCacheTTL = 24 * time.Hour

var pokemonCache *redis.Client

func openPokemonCache(ctx context.Context, redisURL string) *redis.Client {
	if strings.TrimSpace(redisURL) == "" {
		log.Print("Redis cache disabled: REDIS_URL is not set")
		return nil
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Redis cache disabled: invalid REDIS_URL: %v", err)
		return nil
	}

	client := redis.NewClient(options)
	pingContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(pingContext).Err(); err != nil {
		log.Printf("Redis cache disabled: %v", err)
		client.Close()
		return nil
	}

	log.Print("Redis cache enabled")
	return client
}

func getCachedPokemon(ctx context.Context, name string) (pokemon, bool) {
	if pokemonCache == nil {
		return pokemon{}, false
	}

	key := pokemonCacheKey(name)
	data, err := pokemonCache.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return pokemon{}, false
	}
	if err != nil {
		log.Printf("Redis get %q: %v", key, err)
		return pokemon{}, false
	}

	var cached pokemon
	if err := json.Unmarshal(data, &cached); err != nil {
		log.Printf("Redis decode %q: %v", key, err)
		if err := pokemonCache.Del(ctx, key).Err(); err != nil {
			log.Printf("Redis delete invalid cache entry %q: %v", key, err)
		}
		return pokemon{}, false
	}

	return cached, true
}

func cachePokemon(ctx context.Context, name string, value pokemon) {
	if pokemonCache == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("encode Pokemon cache value: %v", err)
		return
	}

	key := pokemonCacheKey(name)
	if err := pokemonCache.Set(ctx, key, data, pokemonCacheTTL).Err(); err != nil {
		log.Printf("Redis set %q: %v", key, err)
	}
}

func pokemonCacheKey(name string) string {
	return "pokemon:v1:" + strings.ToLower(strings.TrimSpace(name))
}
