// Wails Runtime Shim
// Wails 3 会在运行时自动注入这些函数到全局作用域

// 等待 Wails 运行时初始化
const getRuntime = () => {
  if (typeof window !== 'undefined' && window.wails) {
    return window.wails
  }
  return null
}

export const Call = {
  ByID: (...args) => {
    const runtime = getRuntime()
    if (runtime?.Call?.ByID) {
      return runtime.Call.ByID(...args)
    }
    console.warn('Wails runtime not yet initialized, Call.ByID called with:', args)
    return Promise.resolve()
  }
}

export const Create = {
  ByID: (...args) => {
    const runtime = getRuntime()
    if (runtime?.Create?.ByID) {
      return runtime.Create.ByID(...args)
    }
    console.warn('Wails runtime not yet initialized, Create.ByID called with:', args)
    return {}
  }
}

// Nullable 辅助函数（用于处理可空类型）
export function Nullable(value) {
  return value === null || value === undefined ? null : value
}

// 其他可能需要的运行时函数
export const Events = {
  Emit: (eventName, ...args) => {
    const runtime = getRuntime()
    if (runtime?.Events?.Emit) {
      return runtime.Events.Emit(eventName, ...args)
    }
    console.warn('Wails runtime not yet initialized, Events.Emit called')
  },
  On: (eventName, callback) => {
    const runtime = getRuntime()
    if (runtime?.Events?.On) {
      return runtime.Events.On(eventName, callback)
    }
    console.warn('Wails runtime not yet initialized, Events.On called')
    return () => {}
  }
}
