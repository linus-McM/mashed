export namespace domain {
	
	export class DiffFileStat {
	    path: string;
	    added: number;
	    removed: number;
	    isBinary: boolean;
	    isNew: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiffFileStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.added = source["added"];
	        this.removed = source["removed"];
	        this.isBinary = source["isBinary"];
	        this.isNew = source["isNew"];
	    }
	}
	export class NotificationEvent {
	    id: string;
	    agentId: string;
	    agentName: string;
	    model: string;
	    repoName: string;
	    repoPath: string;
	    repoBranch: string;
	    eventType: string;
	    summary: string;
	    // Go type: time
	    timestamp: any;
	    read: boolean;
	    priority: number;
	    tokensUsed: number;
	    tokensMax: number;
	    tmuxTarget: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new NotificationEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.agentId = source["agentId"];
	        this.agentName = source["agentName"];
	        this.model = source["model"];
	        this.repoName = source["repoName"];
	        this.repoPath = source["repoPath"];
	        this.repoBranch = source["repoBranch"];
	        this.eventType = source["eventType"];
	        this.summary = source["summary"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.read = source["read"];
	        this.priority = source["priority"];
	        this.tokensUsed = source["tokensUsed"];
	        this.tokensMax = source["tokensMax"];
	        this.tmuxTarget = source["tmuxTarget"];
	        this.pid = source["pid"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScopedDiff {
	    files: DiffFileStat[];
	
	    static createFrom(source: any = {}) {
	        return new ScopedDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], DiffFileStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorktreeInfo {
	    path: string;
	    branch: string;
	    isOrphaned: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorktreeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.branch = source["branch"];
	        this.isOrphaned = source["isOrphaned"];
	    }
	}

}

