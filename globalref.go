package jni

// GlobalRef is an Object that has been promoted to a JNI global reference.
// It has the same layout as Object so that it can be cast to *Class or
// other reference types via unsafe.Pointer when needed.
type GlobalRef = Object
