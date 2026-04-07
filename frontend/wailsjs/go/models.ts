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
	export class LogLine {
	    kind: string;
	    text: string;
	    // Go type: time
	    ts: any;
	
	    static createFrom(source: any = {}) {
	        return new LogLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.text = source["text"];
	        this.ts = this.convertValues(source["ts"], null);
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
	    isSubAgent: boolean;
	    parentAgentId?: string;
	    subAgentName?: string;
	    subAgentDesc?: string;
	    subAgentStatus?: string;
	    subAgentResult?: string;
	    subAgentLogLines?: LogLine[];
	
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
	        this.isSubAgent = source["isSubAgent"];
	        this.parentAgentId = source["parentAgentId"];
	        this.subAgentName = source["subAgentName"];
	        this.subAgentDesc = source["subAgentDesc"];
	        this.subAgentStatus = source["subAgentStatus"];
	        this.subAgentResult = source["subAgentResult"];
	        this.subAgentLogLines = this.convertValues(source["subAgentLogLines"], LogLine);
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

export namespace main {
	
	export class LocalFontFile {
	    fileName: string;
	    weight: string;
	    style: string;
	    format: string;
	    base64: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalFontFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileName = source["fileName"];
	        this.weight = source["weight"];
	        this.style = source["style"];
	        this.format = source["format"];
	        this.base64 = source["base64"];
	    }
	}
	export class LocalFontFamily {
	    family: string;
	    files: LocalFontFile[];
	
	    static createFrom(source: any = {}) {
	        return new LocalFontFamily(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.family = source["family"];
	        this.files = this.convertValues(source["files"], LocalFontFile);
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
	
	export class NerdFontEntry {
	    family: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new NerdFontEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.family = source["family"];
	        this.filePath = source["filePath"];
	    }
	}
	export class VSCodeThemeEntry {
	    label: string;
	    extensionId: string;
	    themePath: string;
	    uiTheme: string;
	
	    static createFrom(source: any = {}) {
	        return new VSCodeThemeEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.extensionId = source["extensionId"];
	        this.themePath = source["themePath"];
	        this.uiTheme = source["uiTheme"];
	    }
	}
	export class mashedConfig {
	    devDir: string;
	    theme?: string;
	    vscodiumExtPath?: string;
	    importedTheme?: string;
	    monoFont?: string;
	    fontSize?: number;
	
	    static createFrom(source: any = {}) {
	        return new mashedConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.devDir = source["devDir"];
	        this.theme = source["theme"];
	        this.vscodiumExtPath = source["vscodiumExtPath"];
	        this.importedTheme = source["importedTheme"];
	        this.monoFont = source["monoFont"];
	        this.fontSize = source["fontSize"];
	    }
	}

}

