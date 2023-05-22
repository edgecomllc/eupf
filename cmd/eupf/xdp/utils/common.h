#pragma once

#define increment_counter( OBJECT, COUNTER)   \
    if (OBJECT)                          \
        __sync_fetch_and_add(&OBJECT->COUNTER, 1);