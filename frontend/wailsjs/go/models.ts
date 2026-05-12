export namespace core {
	
	export class AudioNode {
	    id: string;
	    name: string;
	    type: string;
	    is_default: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AudioNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.is_default = source["is_default"];
	    }
	}
	export class Device {
	    id: string;
	    name: string;
	    ip: string;
	    platform: string;
	    is_local: boolean;
	    last_seen: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.platform = source["platform"];
	        this.is_local = source["is_local"];
	        this.last_seen = source["last_seen"];
	    }
	}

}

