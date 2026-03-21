package jni

// ProxyHandlerFull is like ProxyHandler but also receives the
// java.lang.reflect.Method object for return type inspection
// (e.g. detecting void callbacks via method.getReturnType()).
type ProxyHandlerFull func(env *Env, method *Object, methodName string, args []*Object) (*Object, error)
