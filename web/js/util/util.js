let util = {};

util.def_fill = function (def, inst) {
  for (let key in def) {
    if (!key) {
      continue;
    }
    let val = inst[key];
    let val_def = def[key];
    let val_type = typeof val;
    let val_def_type = typeof val_def;
    if (typeof val === 'undefined') {
      if (val_type !== val_def_type) {
        if (val_type === 'number' && val_def_type === 'string') {
          val_def = val + '';
        } else if (val_type === 'string' && val_def_type === 'number') {
          val = val * 1;
          if (!isNaN(val)) {
            val_def = val;
          }
        }
      }
      inst[key] = val_def;
    } else {
      if (val instanceof Object) {
        util.def_fill(val_def, val);
      }
    }
  }
  return inst;
};

util.obj_clone_shadow = function (src, des) {
  if (!des) {
    des = {};
  }
  for (let k in src) {
    des[k] = src[k];
  }
  return des;
};

util.time_diff = function (late, early) {
  return late.getTime() - early.getTime();
};

util.time_hhmmss = function (time) {
  if (!time) {
    time = new Date();
  }
  let hour = time.getHours();
  let minute = time.getMinutes();
  let second = time.getSeconds();
  if (hour < 10) hour = '0' + hour;
  if (minute < 10) minute = '0' + minute;
  if (second < 10) second = '0' + second;
  return hour + ':' + minute + ':' + second;
};

