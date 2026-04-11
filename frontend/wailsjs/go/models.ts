export namespace main {
	
	export class AppConfig {
	    screenshotsDirPath: string;
	    apiKey: string;
	    hideOnClose: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.screenshotsDirPath = source["screenshotsDirPath"];
	        this.apiKey = source["apiKey"];
	        this.hideOnClose = source["hideOnClose"];
	    }
	}

}

