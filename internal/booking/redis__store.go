package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb}
}

func sessionKey(id string) string {
	return fmt.Sprintf("session %s", id)
}

func (s *RedisStore) Book(b Booking) (Booking, error) {
	session, err := s.hold(b)
	if err != nil {
		return Booking{}, err
	}
	log.Printf("session booked %v", session)
	return session, nil
}

func (s *RedisStore) ListBookings(movieID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", movieID)
	var sessions []Booking

	ctx := context.Background()

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		session, err := parseSession(val)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	if err := iter.Err(); err != nil {
		log.Println(err)
	}

	return sessions
}

func parseSession(val string) (Booking, error) {
	var data Booking
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}
	return data, nil
}

func (s *RedisStore) Confirm(ctx context.Context, sessionID string, userID string) (Booking, error) {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return Booking{}, err
	}

	session.Status = "confirmed"
	session.ExpiresAt = time.Time{}
	val, err := json.Marshal(session)
	if err != nil {
		return Booking{}, err
	}
	if err := s.rdb.Set(ctx, sk, val, 0).Err(); err != nil {
		return Booking{}, err
	}
	if err := s.rdb.Persist(ctx, sessionKey(sessionID)).Err(); err != nil {
		return Booking{}, err
	}

	return session, nil
}

func (s *RedisStore) getSession(ctx context.Context, sessionID string, userID string) (Booking, string, error) {
	sk, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return Booking{}, "", err
	}

	val, err := s.rdb.Get(ctx, sk).Result()
	if err != nil {
		return Booking{}, "", err
	}

	session, err := parseSession(val)
	if err != nil {
		return Booking{}, "", err
	}
	if session.UserID != userID {
		return Booking{}, "", ErrSessionDoesNotBelongUser
	}

	return session, sk, nil
}

func (s *RedisStore) Release(ctx context.Context, sessionID string, userID string) error {
	session, sk, err := s.getSession(ctx, sessionID, userID)
	if err != nil {
		return err
	}
	if session.Status == "confirmed" {
		return ErrConfirmedCannotRelease
	}

	return s.rdb.Del(ctx, sk, sessionKey(sessionID)).Err()
}

func (s *RedisStore) hold(b Booking) (Booking, error) {
	id := uuid.New().String()
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)

	// setting the Values
	b.ID = id
	b.Status = "held"
	b.ExpiresAt = now.Add(defaultHoldTTL)
	val, err := json.Marshal(b)
	if err != nil {
		return Booking{}, err
	}

	// redis
	res := s.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX",
		TTL:  defaultHoldTTL,
	})
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return Booking{}, ErrSeatAlreadyBooked
		}
		return Booking{}, err
	}

	// check whether the booking goes through, return the seat already booked error if it doesn't
	ok := res.Val() == "OK"
	if !ok {
		return Booking{}, ErrSeatAlreadyBooked
	}

	if err := s.rdb.Set(ctx, sessionKey(id), key, defaultHoldTTL).Err(); err != nil {
		s.rdb.Del(ctx, key)
		return Booking{}, err
	}
	return b, nil
}
