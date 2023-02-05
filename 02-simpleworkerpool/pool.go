package workerpool

import (
	"errors"
	"log"
	"sync"
)

const (
	defaultCapacity = 100
	maxCapacity     = 10000
)

var (
	ErrNoIdleWorkerInPool = errors.New("no idle worker in pool") // workerpool中任务已满，没有空闲goroutine用于处理新任务
	ErrWorkerPoolFreed    = errors.New("workerpool freed")       // workerpool已终止运行
)

type Pool struct {
	capacity int  // workerpool大小
	preAlloc bool // 是否在创建pool的时候，就预创建workers，默认值为：false

	// 当pool满的情况下，新的Schedule调用是否阻塞当前goroutine。默认值：true
	// 如果block = false，则Schedule返回ErrNoWorkerAvailInPool
	block  bool
	active chan struct{} // active channel

	tasks chan Task // task channel

	wg   sync.WaitGroup // pool销毁时等待所worker退出
	quit chan struct{}  // 通知各个worker退出的信号量
}

type Task func()

// New 创建并运行线程池
func New(capacity int, opts ...Option) *Pool {
	if capacity < 0 {
		capacity = defaultCapacity
	}
	if capacity > maxCapacity {
		capacity = maxCapacity
	}

	p := &Pool{
		capacity: capacity,
		active:   make(chan struct{}, capacity),
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
	}

	// 设置自定义选项
	for _, opt := range opts {
		opt(p)
	}

	log.Printf("workerpool start(preAlloc=%t)\n", p.preAlloc)
	if p.preAlloc { // 初始化时创建workers
		for i := 0; i < p.capacity; i++ {
			p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}

	go p.run() // 运行

	return p
}
func (p *Pool) returnTask(t Task) {
	go func() {
		p.tasks <- t
	}()
}

func (p *Pool) run() {
	idx := len(p.active)

	if !p.preAlloc { // 初始化没有创建workers
	loop:
		for t := range p.tasks {
			// 根据task创建worker, 因为需去除所以再将t放回以使worker处理
			p.returnTask(t)
			select {
			case <-p.quit:
				return
			case p.active <- struct{}{}:
				idx++
				p.newWorker(idx)
			default:
				break loop
			}
		}

	}

	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}:
			idx++
			p.newWorker(idx) // 创建新的协程worker
		}
	}
}

func (p *Pool) newWorker(i int) {
	p.wg.Add(1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("worker[%03d]: recover panic[%s] add exit\n", i, err)
				<-p.active
			}
			p.wg.Done()
		}()

		log.Printf("worker[%03d]: start\n", i)

		for {
			select {
			case <-p.quit:
				log.Printf("worker[%03d]: exit", i)
				<-p.active
				return
			case p := <-p.tasks:
				log.Printf("worker[%03d]: receive a task\n", i)
				p()
			}
		}
	}()
}

func (p *Pool) Schedule(t Task) error {

	select {
	case <-p.quit:
		return ErrWorkerPoolFreed
	case p.tasks <- t:
		return nil
	default:
		if p.block {
			p.tasks <- t
			return nil
		}
		return ErrNoIdleWorkerInPool
	}
}

func (p *Pool) Free() {
	close(p.quit)
	p.wg.Wait()
	log.Printf("workerpool freed(preAlloc=%t)\n", p.preAlloc)
}
