Vue.use(TreeView);
// Vue.component('tree-view', window.TreeView);

// Vue.use(VueBootstrapTypeahead);
Vue.component('vue-bootstrap-typeahead', VueBootstrapTypeahead);


let config_def = {
    socket: {
        type: TEST,
        reconnect: 3,
        test: {
            url: '',
        },
        game: {
            url: ''
        },
        platform: {
            url: 'ws://127.0.0.1:8888'
        }
    },

    call: {
        method: 'ag.echo',
        params: JSON.stringify({
            'hello': 'world',
            'oops': 'O_O!',
            'doc': 'click button call to invoke vertx.eventbus.send()'
        }, null, '  '),
        script: '// click button exec to run lua script; \n return "hello world from lua" + oops;',
        timeout: 30,
        gatekey: "",
        method_map: {},
        method_list: [],
        method_list_limit: 256,
        pretty_limit: 1024,
        record_limit: 12,
        request_json_depth: 2,
        response_json_depth: 2,
        response_json_pretty: true
    },

    editor: {
        params_height: 300,
        script_height: 300
    },

    info: {
        limit: 22
    },

    timer: {
        type: 'call',
        active: false,
        interval: 5000
    },

    game: {
        id: 5,
        type: 'slot',
        operator_id: 1,
        username: 'aspect_test9',
        password: 'test',
        bet_multiplier: 20,
        emulation: 'main.stops=12,1,12,10,12,9,2,1,10,1,4,10,1,8,1,10,0,1,6,1',
        progressive: false,
        progressive_magnifier: 1
    }
};

let data = {
    msgs: [],
    socket: null,
    socket_pending: [],
    socket_reconnect_timer: null,
    records: [],
    record_id: 0,
    request: [],
    response: [],
    session: {
        balance: 0
    },
    config: util.obj_clone_shadow(config_def)
};

let methods = {

    config_load() {
        let config = window.localStorage.getItem(window.location.pathname);
        if (config) {
            config = JSON.parse(config);
            for (let k in config) {
                let v = config[k];
                if (typeof v !== 'undefined') {
                    this.config[k] = v;
                }
            }
        }
        util.def_fill(config_def, this.config);
        console.log('[config.load]', this.config);
    },

    config_save(print) {
        let persist = {};
        let config = this.$data.config;
        for (let k in config) {
            let v = config[k];
            if (typeof v === 'function') {
                continue;
            }
            persist[k] = v;
        }
        if (print) {
            console.log('[config.save]', persist);
        }
        persist = JSON.stringify(persist);
        window.localStorage.setItem(window.location.pathname, persist);
    },

    info(msg, color, to_console) {
        let time = util.time_hhmmss();
        this.msgs.unshift({
            msg: msg,
            time: time,
            color: color || 'grey'
        });
        if (this.msgs.length > this.config.info.limit) {
            this.msgs.pop();
        }
        if (typeof to_console !== 'undefined' && to_console) {
            console.log(msg);
        }
    },

    error(ex) {
        let msg = ex.stack ? ex.stack : ex;
        console.error(ex);
        return this.info(msg, 'red', false);
    },

    type(type) {
        if (type) {
            this.config.socket.type = type;
        }
        return this.config.socket.type;
    },

    url(url) {
        let type = this.type();
        if (url) {
            this.config.socket[type].url = url;
        }
        return this.config.socket[type].url;
    },

    connect() {
        if (this.socket) {
            console.error('[socket] already connected');
            return;
        }

        let type = this.type();
        let url = this.url();

        if (!url) {
            if (type === TEST) {
                url = 'ws' + window.location.origin.replace('https', '').replace('http', '');
            } else {
                url = 'ws://127.0.0.1:8888';
            }
        }
        if (!url) {
            this.error('no url is set');
            return;
        }
        this.url(url);
        this.socket = new WebSocket(url);
        this.socket.onopen = this.ws_onopen;
        this.socket.onmessage = this.ws_onmessage;
        this.socket.onerror = this.ws_onerror;
        this.socket.onclose = this.ws_onclose;
    },

    reconnect(type) {
        if (type) {
            this.type(type);
        }
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
        this.config_save();
        this.connect();
    },

    ws_onopen() {
        let type = this.type();
        this.info('connected to ' + this.url() + ' | ' + type);
        if (this.socket_reconnect_timer) {
            clearInterval(this.socket_reconnect_timer);
            this.socket_reconnect_timer = null;
        }
        this.record_id = 0;
        if (type === PLATFORM) {
            this.authenticate();
        }
    },

    ws_onmessage(response) {
        let packet = response.data;
        try {
            packet = JSON.parse(packet);
        } catch (ex) {

        }
        let target;
        let id = packet.id;
        let n = this.records.length;
        if (id && id > 0) {
            for (let i = 0; i < n; i++) {
                let record = this.records[i];
                if (record.id === packet.id) {
                    target = record;
                    break;
                }
            }
        } else {
            this.handle_push(packet);
            return;
        }

        if (target) {
            this.response[id] = packet;
            target.pretty = false;
            target.response = packet;
            target.response_time = new Date();
            target.consume = packet.consume;

            let r = target.response.result = target.response.result || target.response.error;

            if (!r) {
                r = target.response = {result: packet};
            }

            if (r && typeof r === 'string') {
                r = r.trim();
                if (r.charAt(0) === '{' || r.charAt(0) === '[') {
                    try {
                        r = JSON.parse(r);
                        target.response.result = r;
                        target.pretty = this.config.call.response_json_pretty;
                    } catch (ex) {
                        this.error(ex);
                    }
                }
            }

            if (r.result && typeof r.result === 'string') {
                r.result = r.result.trim();
                if (r.result.charAt(0) === '{' || r.result.charAt(0) === '[') {
                    try {
                        r.result = JSON.parse(r.result);
                        target.pretty = this.config.call.response_json_pretty;
                    } catch (ex) {
                    }
                }
            }

            if (typeof target.consume === 'undefined') {
                target.consume = target.response_time.getTime() - target.request_time.getTime();
            }

        }

        console.log('[response] ', id, packet);
        // this.handle_game_response(this.request[id], packet);
        this.$forceUpdate();
    },

    handle_push(packet) {

        let msg;
        try {
            msg = JSON.parse(packet.result);
        } catch (ex) {
            msg = {
                source: '',
                value: packet.result
            };
        }

        this.records.unshift({
            id: packet.id || -1,
            request: {
                action: 'push',
                params: {
                    source: msg.source
                }
            },
            response: {
                result: msg.value
            },
            pretty: true,
            consume: 0
        });
        if (this.records.length > this.config.call.record_limit) {
            this.records.pop();
        }
        console.log('[push] ', msg);
        this.$forceUpdate();
    },

    ws_send(packet, delay, remember) {

        if (delay) {
            setTimeout(function () {
                this.ws_send(packet);
            }.bind(this), delay);
            return;
        }

        if (!this.socket) {
            this.socket_pending.unshift(packet);
            this.connect();
            return;
        }

        if (typeof packet === 'string') {
            packet = JSON.parse(packet);
        }

        let id = packet.id = ++this.record_id;
        this.records.unshift({
            id: id,
            request: packet,
            request_time: new Date()
        });

        this.request[id] = packet;
        if (this.records.length > this.config.call.record_limit) {
            this.records.pop();
        }

        if (remember) {
            if (packet.method) {
                this.config.call.method = packet.method;
            }
            if (packet.params) {
                this.config.call.params = JSON.stringify(packet.params, null, '  ');
            }
        }

        console.log('[request]', id, packet);
        let string = JSON.stringify(packet);
        this.socket.send(string);
    },

    ws_onerror(ex) {
        this.error(ex);

    },

    ws_onclose() {
        this.socket = null;
        this.error('disconnected from ' + this.url());
        if (!this.config.socket.reconnect || this.config.socket.reconnect <= 0) {
            this.config.socket.reconnect = 5;
        }
        if (this.socket_reconnect_timer) {
            clearInterval(this.socket_reconnect_timer);
        }

        this.socket_reconnect_timer = setInterval(function () {
            this.connect();
        }.bind(this), this.config.socket.reconnect * 1000);
    },

    select(method) {

        if (method) {
            this.config.call.method = method;
        } else {
            method = this.config.call.method;
        }
        // this.$refs.call_method_typeahead.inputValue = method;
        if (!this.config.call.method_map) {
            this.config.call.method_map = {};
        }
        let box = this.config.call.method_map[method];
        if (typeof box === 'string') {
            box = {params: box, script: ''};
        }
        this.config.call.params = box.params;
        this.config.call.script = box.script;
        this.editor_params.setValue(this.config.call.params);
        this.editor_script.setValue(this.config.call.script);
        this.editor_params.clearSelection();
        this.editor_script.clearSelection();
    },

    method_list_delete_item(method) {
        this.config.call.method_map[method] = null;
        let method_list = this.config.call.method_list;
        let new_list = [];
        for (let i = 0; i < method_list.length; ++i) {
            let one = method_list[i];
            if (one !== method) {
                new_list.push(one);
            }
        }
        this.config.call.method_list = new_list;
    },

    record(type) {
        let method = this.config.call.method;
        let method_list = this.config.call.method_list;
        let new_list = [method];
        for (let i = 0; i < method_list.length && i < this.config.call.method_list_limit; i++) {
            let one = method_list[i];
            if (one !== method) {
                new_list.push(one);
            }
        }
        this.config.call.method_list = new_list;
        this.config.call.method_map[method] = {
            type: type,
            params: this.config.call.params,
            script: this.config.call.script
        };
    },

    record_clear() {
        this.records.splice(0, this.records.length);
    },

    get_call_params() {
        let json;
        try {
            json = JSON.parse(this.config.call.params);
            this.config.call.params = JSON.stringify(json, null, '  ');
        } catch (ex) {
            json = this.config.call.params;
        }
        return json;
    },

    get_call_script() {
        return this.config.call.script;
    },

    call(extra) {

        let params = this.get_call_params();
        let method = this.config.call.method;
        let gatekey = this.config.call.gatekey;

        this.record('call');
        this.config_save();

        let packet = {
            action: 'call',
            method: method,
            params: params,
            gatekey: gatekey,
            timeout: (this.config.call.timeout || 60) * 1000
        };

        if (extra) {
            for (let k in extra) {
                packet[k] = extra[k]
            }
        }

        this.ws_send(packet);
    },

    exec() {
        let params = this.get_call_params();
        let script = this.get_call_script();
        let name = this.config.call.method;

        this.record('exec');
        this.config_save();

        this.ws_send({
            action: 'script',
            name: name,
            params: params,
            script: script,
            timeout: (this.config.call.timeout || 60) * 1000,
        });
    },

    deploy() {
        let params = this.get_call_params();
        let script = this.get_call_script();
        let name = this.config.call.method;

        this.record('deploy');
        this.config_save();

        this.ws_send({
            action: 'deploy',
            name: name,
            params: params,
            script: script,
            timeout: (this.config.call.timeout || 60) * 1000,
        });
    },

    /* ============================= request routine ============================== */

    game_type() {
        return this.config.game.type || 'slot';
    },

    handle_game_response(request, response) {
        if (this.type() !== PLATFORM) {
            return;
        }
        try {
            let method = request.method || response.method;
            let game_result = response.result;
            switch (method) {
                case 'update':
                    let balance = response.params.balance;
                    if (balance) {
                        this.session.balance = balance;
                    }
                    break;
                case METHOD_AUTHENTICATE:
                    if (game_result && game_result.currency) {
                        this.session.currency = game_result.currency;
                        this.session.language = game_result.language;
                        this.info('authenticate success ');
                        this.register(50);
                    } else {
                        this.error('authenticate fail');
                    }
                    break;
                case GAME_METHOD.GameRegister:
                    this.info('slot game registered @ ' + this.url());
                    break;
            }
            if (game_result && game_result.state) {
                this.session.balance = game_result.balance / 100;
            }
        } catch (ex) {
            this.error(ex);
        }
    },

    authenticate() {
        let type = this.type();
        if (type !== PLATFORM) {
            return;
        }
        let game = this.config.game;
        this.ws_send({
            'method': METHOD_AUTHENTICATE,
            'params': {
                'gameId': game.id,
                'operatorId': game.operator_id,
                'user': game.username,
                'password': game.password
            }
        }, 0);
    },

    register(delay) {
        let method = this.game_type() + 'GameRegister';
        this.ws_send({
            method: method
        }, delay);
    },

    play(delay, params_def, emulation) {
        let type = this.game_type();
        let game = this.config.game;
        let bet_multiplier = (game.bet_multiplier || 1) * 1;
        let method = type + 'GamePlay';
        let params = {
            betMultiplier: bet_multiplier,
        };
        if (emulation) {
            params.emulation = emulation;
        }
        if (game.progressive_magnifier && game.progressive_magnifier > 0) {
            params.progressive = {};
            if (game.progressive_magnifier > 0) {
                params.progressive.emulation = {
                    magnifier: (game.progressive_magnifier || 1) * 1
                };
            }

        }
        if (params_def) {
            util.def_fill(params_def, params);
        }
        this.ws_send({
            method: method,
            params: params
        }, delay, true);
    },

    timer_invoke() {

        if (this.config.timer.active) {
            window.clearInterval(this.timer);
            this.config.timer.active = false;
            this.timer = null;
        } else {
            this.timer = window.setInterval(function () {
                this.call();
            }.bind(this), this.config.timer.interval);
            this.config.timer.active = true;
        }

    },

    editor_init() {
        if (!this.editor_params) {
            this.editor_params = ace.edit('editor_params', {
                mode: 'ace/mode/json',
                selectionStyle: 'text',
                highlightActiveLine: true,
                highlightSelectedWord: true,
                cursorStyle: 'ace',
                newLineMode: 'unix',
                fontSize: '0.8em'
            });
            this.editor_params.setOption('wrap', 'free');
            this.editor_params.setTheme('ace/theme/github');
            this.editor_params.setValue(this.config.call.params);
            this.editor_params.clearSelection();
            this.editor_params.getSession().on('change', function () {
                this.config.call.params = this.editor_params.getValue().trim();
            }.bind(this));
        }
        if (!this.editor_script) {
            this.editor_script = ace.edit('editor_script', {
                mode: 'ace/mode/lua',
                selectionStyle: 'text',
                highlightActiveLine: true,
                highlightSelectedWord: true,
                cursorStyle: 'ace',
                newLineMode: 'unix',
                fontSize: '0.8em'
            });
            this.editor_script.setOption('wrap', 'free');
            this.editor_script.setTheme('ace/theme/github');
            this.editor_script.setValue(this.config.call.script);
            this.editor_script.clearSelection();
            this.editor_script.getSession().on('change', function () {
                this.config.call.script = this.editor_script.getValue().trim();
            }.bind(this));
        }
        this.editor_params.setFontSize('0.75em');
        this.editor_script.setFontSize('0.75em');
        jQuery('#editor_params').css('height', this.config.editor.params_height + 'px');
        jQuery('#editor_script').css('height', this.config.editor.script_height + 'px');
        this.editor_params.resize();
        this.editor_script.resize();
    },


};

window.vue = new Vue({
    el: '#vcontainer',
    data: data,
    methods: methods,
    watch: {
        'config.call.method': function (nvalue, ovalue) {
            this.$refs.call_method_typeahead.inputValue = nvalue;
        }
    },
    mounted: function () {
        this.config_load();
        this.editor_init();
        this.$refs.call_method_typeahead.inputValue = this.config.call.method;
        this.select(this.config.call.method);
        this.connect();
    }
});

let title = window.location.pathname;
title = title.replace('/', ' ');
title = title.replace('/', ' ');
title = title.replace('virtual', '');
title = title.replace('.html', '');
window.document.title = title;