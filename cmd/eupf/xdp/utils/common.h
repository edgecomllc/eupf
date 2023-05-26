#pragma once

#define increment_counter(OBJECT, COUNTER) \
        OBJECT->COUNTER++;

#define increment_counter_sync(OBJECT, COUNTER) \
        __sync_fetch_and_add(&OBJECT->COUNTER, 1);
