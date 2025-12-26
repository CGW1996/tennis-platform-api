package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"tennis-platform/backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客戶端結構
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient 創建新的 Redis 客戶端
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	// 創建 Redis 客戶端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")

	return &RedisClient{Client: rdb}, nil
}

// Close 關閉 Redis 連接
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// GetClient 獲取 Redis 客戶端實例
func (r *RedisClient) GetClient() *redis.Client {
	return r.Client
}

// HealthCheck 檢查 Redis 連接健康狀態
func (r *RedisClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}

// Set 設置鍵值對
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get 獲取值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del 刪除鍵
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists 檢查鍵是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// Expire 設置鍵的過期時間
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// HSet 設置哈希字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet 獲取哈希字段值
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll 獲取哈希所有字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// HDel 刪除哈希字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.Client.HDel(ctx, key, fields...).Err()
}

// SAdd 添加集合成員
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SAdd(ctx, key, members...).Err()
}

// SMembers 獲取集合所有成員
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

// SRem 移除集合成員
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.SRem(ctx, key, members...).Err()
}

// ZAdd 添加有序集合成員
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.Client.ZAdd(ctx, key, members...).Err()
}

// ZRange 獲取有序集合範圍內的成員
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore 根據分數範圍獲取有序集合成員
func (r *RedisClient) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.Client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRem 移除有序集合成員
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.Client.ZRem(ctx, key, members...).Err()
}

// Publish 發布消息到頻道
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe 訂閱頻道
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}

// LPush 從列表左側推入元素
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

// RPush 從列表右側推入元素
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.RPush(ctx, key, values...).Err()
}

// LPop 從列表左側彈出元素
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.Client.LPop(ctx, key).Result()
}

// RPop 從列表右側彈出元素
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

// LRange 獲取列表範圍內的元素
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.LRange(ctx, key, start, stop).Result()
}

// LLen 獲取列表長度
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.Client.LLen(ctx, key).Result()
}
