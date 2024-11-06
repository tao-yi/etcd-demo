package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type DistributedLock struct {
	Key        string
	Value      string
	LeaseID    clientv3.LeaseID
	etcdClient *clientv3.Client
}

func (dl *DistributedLock) Lock(ctx context.Context, ttl int64) error {
	lease, err := dl.etcdClient.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	_, err = dl.etcdClient.Put(ctx, dl.Key, dl.Value, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	dl.LeaseID = lease.ID
	log.Printf("Lock acquired: %s", dl.Key)
	return nil
}

func (dl *DistributedLock) Unlock(ctx context.Context) error {
	_, err := dl.etcdClient.Delete(ctx, dl.Key)
	if err != nil {
		return err
	}

	_, err = dl.etcdClient.Revoke(ctx, dl.LeaseID)
	if err != nil {
		return err
	}

	log.Printf("Lock released: %s", dl.Key)
	return nil
}

func main() {
	endpoints := []string{"localhost:2379"}

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		fmt.Printf("Error connecting to etcd: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()
	lockKey := "my-lock"
	lockValue := "my-value"

	dl := DistributedLock{
		Key:        lockKey,
		Value:      lockValue,
		etcdClient: client,
	}

	err = dl.Lock(ctx, 10)
	if err != nil {
		fmt.Printf("Error acquiring lock: %v", err)
		os.Exit(1)
	}

	// Simulate a critical section
	time.Sleep(5 * time.Second)

	err = dl.Unlock(ctx)
	if err != nil {
		fmt.Printf("Error releasing lock: %v", err)
		os.Exit(1)
	}
}
