// Wails Runtime Shim
// Wails 3 会在运行时自动注入这些函数到全局作用域

export const Call = window.wails?.Call || {
  ByID: (...args) => {
    console.warn('Wails runtime not yet initialized')
    return Promise.resolve()
  }
}

export const Create = window.wails?.Create || {
  ByID: (...args) => {
    console.warn('Wails runtime not yet initialized')
    return {}
  }
}
