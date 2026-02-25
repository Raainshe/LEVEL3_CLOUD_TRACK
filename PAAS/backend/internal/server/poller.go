package server

import (
	"context"
	"log"
	"strings"
	"time"

	"backend/internal/kube"
	"backend/internal/models"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	watchBackoffMax   = 30 * time.Second
	watchBackoffInit  = 500 * time.Millisecond
	statusPollSeconds = 30
)

func (s *Server) listRedisFailoversByNamespace(ctx context.Context) (*unstructured.UnstructuredList, error) {
	names, err := kube.ListNamespaceNames(ctx, s.kubeClient)
	if err != nil {
		return nil, err
	}
	out := &unstructured.UnstructuredList{}
	for _, ns := range names {
		list, err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(ns).List(ctx, v1.ListOptions{})
		if err != nil {
			continue
		}
		out.Items = append(out.Items, list.Items...)
	}
	return out, nil
}

func (s *Server) processInstanceStatus(ctx context.Context, item *unstructured.Unstructured) (wrote bool, err error) {
	var instance models.RedisInstance
	instance.ConvertUnstructuredToRedisInstace(item)
	liveStatus := kube.GetStatusFromStatefulSets(ctx, s.kubeClient, instance.Namespace, instance.Name, instance.RedisReplicas, instance.SentinelReplicas)
	if liveStatus != "" {
		instance.Status = liveStatus
	} else if instance.Status == "Unknown" || instance.Status == "" || instance.Status == "-" {
		instance.Status = "Provisioning"
	}
	if instance.Status == "" {
		instance.Status = "Provisioning"
	}

	cached, err := s.db.GetInstanceStatusCache(ctx, instance.Name, instance.Namespace)
	if err != nil {
		return false, err
	}

	current := instance.Status
	if cached == "" {
		msg := "Instance first seen: " + current
		svcLog := &models.ServiceLog{
			InstanceName: instance.Name,
			Namespace:    instance.Namespace,
			EventType:    "status_change",
			FromStatus:   "",
			ToStatus:     current,
			Message:      msg,
			Timestamp:    time.Now(),
		}
		if err := s.db.InsertServiceLog(ctx, svcLog); err != nil {
			return false, err
		}
		_ = s.db.SetInstanceStatusCache(ctx, instance.Name, instance.Namespace, current)
		return true, nil
	}

	if cached == current {
		return false, nil
	}

	eventType := "status_change"
	if current == "Failed" {
		eventType = "failure"
	}
	msg := "Status changed from " + cached + " to " + current
	svcLog := &models.ServiceLog{
		InstanceName: instance.Name,
		Namespace:    instance.Namespace,
		EventType:    eventType,
		FromStatus:   cached,
		ToStatus:     current,
		Message:      msg,
		Timestamp:    time.Now(),
	}
	if err := s.db.InsertServiceLog(ctx, svcLog); err != nil {
		return false, err
	}
	_ = s.db.SetInstanceStatusCache(ctx, instance.Name, instance.Namespace, current)
	return true, nil
}

func (s *Server) RunStatusWatcher(ctx context.Context) {
	backoff := watchBackoffInit
	for {
		if ctx.Err() != nil {
			return
		}
		err := s.runStatusWatcherOnce(ctx)
		if err != nil {
			log.Printf("[service-log watcher] watch ended: %v; reconnecting in %v", err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				if backoff < watchBackoffMax {
					backoff *= 2
					if backoff > watchBackoffMax {
						backoff = watchBackoffMax
					}
				}
			}
		} else {
			backoff = watchBackoffInit
		}
	}
}

func (s *Server) runStatusWatcherOnce(ctx context.Context) error {
	list, err := s.kubeClient.Resource(kube.RedisFailOver).List(ctx, v1.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "forbidden") {
			list, err = s.listRedisFailoversByNamespace(ctx)
		}
		if err != nil {
			return err
		}
	}
	rv := list.GetResourceVersion()
	if rv == "" {
		rv = "0"
	}

	watcher, err := s.kubeClient.Resource(kube.RedisFailOver).Watch(ctx, v1.ListOptions{ResourceVersion: rv})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	log.Printf("[service-log watcher] watching RedisFailovers from resourceVersion=%s", rv)
	numWritten := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}
			if ev.Type != watch.Added && ev.Type != watch.Modified {
				continue
			}
			u, ok := ev.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			wrote, err := s.processInstanceStatus(ctx, u)
			if err != nil {
				log.Printf("[service-log watcher] process %s/%s: %v", u.GetNamespace(), u.GetName(), err)
				continue
			}
			if wrote {
				numWritten++
			}
		}
	}
}

func (s *Server) RunStatusPoller(ctx context.Context) {
	s.runStatusSyncOnce(ctx)

	go s.RunStatusWatcher(ctx)

	ticker := time.NewTicker(statusPollSeconds * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runStatusSyncOnce(ctx)
		}
	}
}

func (s *Server) runStatusSyncOnce(ctx context.Context) {
	list, err := s.kubeClient.Resource(kube.RedisFailOver).List(ctx, v1.ListOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "forbidden") {
			list, err = s.listRedisFailoversByNamespace(ctx)
			if err != nil {
				log.Printf("[service-log] sync ERROR list by namespace: %v", err)
				return
			}
			log.Printf("[service-log] sync using per-namespace list (%d instances)", len(list.Items))
		} else {
			log.Printf("[service-log] sync ERROR list redis failovers: %v", err)
			return
		}
	}

	numWritten := 0
	for i := range list.Items {
		item := &list.Items[i]
		wrote, err := s.processInstanceStatus(ctx, item)
		if err != nil {
			log.Printf("[service-log] sync process %s/%s: %v", item.GetNamespace(), item.GetName(), err)
			continue
		}
		if wrote {
			numWritten++
		}
	}
	if len(list.Items) > 0 || numWritten > 0 {
		log.Printf("[service-log] sync complete: %d instances, %d logs written", len(list.Items), numWritten)
	}
}
