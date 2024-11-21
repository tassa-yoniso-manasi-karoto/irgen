export namespace main {
	
	export class ProcessParams {
	    url: string;
	    numberOfTitle: number;
	    maxXResolution: number;
	    maxYResolution: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.numberOfTitle = source["numberOfTitle"];
	        this.maxXResolution = source["maxXResolution"];
	        this.maxYResolution = source["maxYResolution"];
	    }
	}

}

