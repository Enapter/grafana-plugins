import _toString from 'lodash/toString';

export const stringify = (v: any) => {
  if (v === null || v === undefined) {
    return '';
  }

  if (v.constructor.name === 'Object' && Object.keys(v).length === 0) {
    return '';
  }

  return v.constructor.name === 'Object' ? JSON.stringify(v) : _toString(v);
};
