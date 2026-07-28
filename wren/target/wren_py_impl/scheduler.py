import heapq
import time
import sys
from collections import deque
from greenlet import greenlet

class Scheduler:
    def __init__(self):
        # FIFO queue for fibers ready to run immediately.
        # Contains tuples of (greenlet, arg, error_message)
        self.ready_queue = deque()
        
        # Min-heap of pending timers.
        # Elements are tuples: (expiry_time, insertion_order, greenlet)
        self.timer_heap = []
        
        # Monotonic time offset for deterministic time tracking
        self.start_time = time.monotonic()
        
        # To break ties in timer heap for identical expiry times (e.g., 0ms)
        self.timer_counter = 0

        # The root schedule loop greenlet
        self.main_fiber = greenlet.getcurrent()
        
        # Tracks the currently executing wren fiber
        self.current_fiber = None
    
    def now(self) -> float:
        """Returns elapsed time in seconds since scheduler startup."""
        return time.monotonic() - self.start_time

    def add(self, fiber_fn) -> greenlet:
        """
        Enqueues a new greenlet directly onto the ready queue's back.
        """
        g = greenlet(fiber_fn)
        self.ready_queue.append((g, None, None))
        return g

    def enqueue(self, fiber: greenlet, arg=None, error=None):
        """
        Appends a fiber to the ready-queue to be resumed.
        """
        self.ready_queue.append((fiber, arg, error))

    def sleep(self, delay_ms: float):
        """
        Suspending the active fiber for a period of time.
        INVARIANT: A 0ms timer MUST NOT run immediately inline. It MUST be placed
        into the timer heap so it queues behind whatever is already pending.
        """
        expiry = self.now() + (delay_ms / 1000.0)
        self.timer_counter += 1
        heapq.heappush(self.timer_heap, (expiry, self.timer_counter, self.current_fiber))
        
        # Yield control back to the scheduler main loop
        self.main_fiber.switch()

    def run_loop(self):
        """
        Main loop: drains ready-queue in strict FIFO order; when empty, advances 
        virtual time to the earliest pending timer, moving due timers to ready-queue.
        Exits when both are empty.
        """
        while self.ready_queue or self.timer_heap:
            # Advance to earliest timer if ready queue is empty
            if not self.ready_queue and self.timer_heap:
                wait_time = self.timer_heap[0][0] - self.now()
                if wait_time > 0:
                    time.sleep(wait_time)
            
            # Move expired timers to ready queue
            current_time = self.now()
            while self.timer_heap and self.timer_heap[0][0] <= current_time:
                _, _, fiber = heapq.heappop(self.timer_heap)
                self.ready_queue.append((fiber, None, None))
                
            # If ready queue has items, dispatch the first one in FIFO
            if self.ready_queue:
                fiber, arg, error = self.ready_queue.popleft()
                self.current_fiber = fiber
                
                try:
                    if error is not None:
                        # Resume fiber with injected error (exception)
                        fiber.throw(RuntimeError(error))
                    else:
                        if not fiber.dead:
                            if arg is None:
                                fiber.switch()
                            else:
                                fiber.switch(arg)
                except SystemExit:
                    # allow deliberate exits to bubble up
                    raise
                except Exception as e:
                    # An uncaught runtime error escaping the fiber forcibly stops loop and exits 70
                    # Standard behavior dictated by "Fiber Resumption" constraint.
                    print(f"{e}", file=sys.stderr)
                    sys.exit(70)
