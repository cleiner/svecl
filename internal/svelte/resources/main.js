// depends on global svelte object

const compile = (source, options) => {
  let result;
  try {
    const svr = svelte.compile(source, options);
    result = {
      code: svr.js.code + '\n//# sourceMappingURL=' + svr.js.map.toUrl(),
      messages: svr.warnings.map(w => ({ type: 'warning', ...w }))
    };
  } catch(e) {
    if (e.name !== 'ParseError' && e.name !== 'ValidationError') {
      throw e;
    }
    result = { code: null, messages: [{ type: 'error', message: e.message, ...e }] };
  }
  return result;
};
