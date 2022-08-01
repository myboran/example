package timewheel

import (
	"container/list"
	"fmt"
	"time"
)

type Job func(interface{})

type TimeWheel struct {
	interval          time.Duration
	ticker            *time.Ticker
	slots             []*list.List        // 双向链表
	timer             map[interface{}]int // 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器, 不会出现并发读写，不加锁直接访问
	currentPos        int                 // 当前指针指向哪一个槽
	slotNum           int                 // 槽数量
	job               Job                 // 定时器回调函数
	addTaskChannel    chan Task           // 新增任务channel
	removeTaskChannel chan interface{}    // 删除任务channel
	stopChannel       chan bool           // 停止定时器channel
	len               int
}

// Task 延时任务
type Task struct {
	delay  time.Duration // 延迟时间
	circle int           // 时间轮需要转动几圈
	key    interface{}   // 定时器唯一标识, 用于删除定时器
	data   interface{}   // 回调函数参数
}

func New(interval time.Duration, slotNum int, job Job) (*TimeWheel, error) {
	if slotNum <= 0 || job == nil {
		return nil, fmt.Errorf("slotNum 必须大于 0, job 不为 nil")
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[interface{}]int),
		currentPos:        0,
		job:               job,
		slotNum:           slotNum,
		addTaskChannel:    make(chan Task),
		removeTaskChannel: make(chan interface{}),
		stopChannel:       make(chan bool),
	}
	tw.initSlots()
	return tw, nil
}

// 初始化槽，每个槽指向一个双向链表
func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

func (tw *TimeWheel) AddTimer(delay time.Duration, key interface{}, data interface{}) error {

	if delay < 0 {
		return fmt.Errorf("delay 必须大于 0")
	}
	_, ok := tw.timer[key]
	if ok {
		return fmt.Errorf("key 已经存在: %v", key)
	}
	tw.addTaskChannel <- Task{delay: delay, key: key, data: data}
	return nil
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识
func (tw *TimeWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
			tw.len++
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for t := l.Front(); t != nil; {
		task := t.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			t = t.Next()
			continue
		}
		// TODO
		tw.len--
		go tw.job(task.data)
		next := t.Next()
		l.Remove(t)
		if task.key != nil {
			delete(tw.timer, task.key)
		}
		t = next
	}
}

func (tw *TimeWheel) addTask(task *Task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	tw.slots[pos].PushBack(task)
	if task.key != nil {
		tw.timer[task.key] = pos
	}
}

func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	circle = delaySeconds / intervalSeconds / tw.slotNum
	pos = (tw.currentPos + delaySeconds/intervalSeconds) % tw.slotNum
	fmt.Println(d, "pos", pos, "circle", circle, "---", tw.currentPos)

	return
}

func (tw *TimeWheel) removeTask(key interface{}) {
	position, ok := tw.timer[key]
	if !ok {
		return
	}
	ts := tw.slots[position]
	for e := ts.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timer, task.key)
			ts.Remove(e)
		}

		e = e.Next()
	}
}
