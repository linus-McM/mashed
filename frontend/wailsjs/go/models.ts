export namespace advice {
	
	export class AdviceMode {
	    name: string;
	    displayName: string;
	    icon: string;
	    order: number;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new AdviceMode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.icon = source["icon"];
	        this.order = source["order"];
	        this.source = source["source"];
	    }
	}

}

export namespace bmad {
	
	export class AgentInfo {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class BmadAgentConfig {
	    id: string;
	    name: string;
	    role: string;
	    persona: string;
	    skills: string[];
	    model: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new BmadAgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.persona = source["persona"];
	        this.skills = source["skills"];
	        this.model = source["model"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class ControlFlowNodeDef {
	    type: string;
	    name: string;
	    description: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new ControlFlowNodeDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	    }
	}
	export class GroupedAgents {
	    bmadAgents: BmadAgentConfig[];
	    localAgents: AgentInfo[];
	    globalAgents: AgentInfo[];
	
	    static createFrom(source: any = {}) {
	        return new GroupedAgents(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bmadAgents = this.convertValues(source["bmadAgents"], BmadAgentConfig);
	        this.localAgents = this.convertValues(source["localAgents"], AgentInfo);
	        this.globalAgents = this.convertValues(source["globalAgents"], AgentInfo);
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
	export class ValidationIssue {
	    field: string;
	    message: string;
	    severity: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidationIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.field = source["field"];
	        this.message = source["message"];
	        this.severity = source["severity"];
	    }
	}
	export class MashedAssetInfo {
	    name: string;
	    path: string;
	    description: string;
	    kind: string;
	    source: string;
	    role: string;
	    completion: string;
	    inputs: string[];
	    outputs: string[];
	    chainable: string;
	    sessionPinned: boolean;
	    issues: ValidationIssue[];
	
	    static createFrom(source: any = {}) {
	        return new MashedAssetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.description = source["description"];
	        this.kind = source["kind"];
	        this.source = source["source"];
	        this.role = source["role"];
	        this.completion = source["completion"];
	        this.inputs = source["inputs"];
	        this.outputs = source["outputs"];
	        this.chainable = source["chainable"];
	        this.sessionPinned = source["sessionPinned"];
	        this.issues = this.convertValues(source["issues"], ValidationIssue);
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
	export class GroupedMashedAssets {
	    localCommands: MashedAssetInfo[];
	    globalCommands: MashedAssetInfo[];
	    localSkills: MashedAssetInfo[];
	    globalSkills: MashedAssetInfo[];
	
	    static createFrom(source: any = {}) {
	        return new GroupedMashedAssets(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.localCommands = this.convertValues(source["localCommands"], MashedAssetInfo);
	        this.globalCommands = this.convertValues(source["globalCommands"], MashedAssetInfo);
	        this.localSkills = this.convertValues(source["localSkills"], MashedAssetInfo);
	        this.globalSkills = this.convertValues(source["globalSkills"], MashedAssetInfo);
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
	export class InputSpec {
	    id: string;
	    source: string;
	    shape?: string;
	    required: boolean;
	    artifactName?: string;
	    upstreamNodeId?: string;
	    prompt?: string;
	    options?: string[];
	    optionsRef?: string;
	    default?: string;
	    validation?: string;
	    maxLength?: number;
	    helpText?: string;
	
	    static createFrom(source: any = {}) {
	        return new InputSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.shape = source["shape"];
	        this.required = source["required"];
	        this.artifactName = source["artifactName"];
	        this.upstreamNodeId = source["upstreamNodeId"];
	        this.prompt = source["prompt"];
	        this.options = source["options"];
	        this.optionsRef = source["optionsRef"];
	        this.default = source["default"];
	        this.validation = source["validation"];
	        this.maxLength = source["maxLength"];
	        this.helpText = source["helpText"];
	    }
	}
	export class IterationGate {
	    kind: string;
	    maxRounds?: number;
	    acceptTokens?: string[];
	    rejectTokens?: string[];
	    customExpr?: string;
	
	    static createFrom(source: any = {}) {
	        return new IterationGate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.maxRounds = source["maxRounds"];
	        this.acceptTokens = source["acceptTokens"];
	        this.rejectTokens = source["rejectTokens"];
	        this.customExpr = source["customExpr"];
	    }
	}
	
	export class ModuleDef {
	    id: string;
	    name: string;
	    version: string;
	    processes: string[];
	    upgradePath?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModuleDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.processes = source["processes"];
	        this.upgradePath = source["upgradePath"];
	    }
	}
	export class NodeInputEntry {
	    inputId: string;
	    round: number;
	    value: string;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new NodeInputEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputId = source["inputId"];
	        this.round = source["round"];
	        this.value = source["value"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class OutputSpec {
	    id: string;
	    target: string;
	    artifactName?: string;
	    description?: string;
	    optional?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OutputSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.target = source["target"];
	        this.artifactName = source["artifactName"];
	        this.description = source["description"];
	        this.optional = source["optional"];
	    }
	}
	export class PendingPrompt {
	    nodeId: string;
	    inputId: string;
	    prompt: string;
	    shape: string;
	    options?: string[];
	    round: number;
	    createdAt: number;
	    promptId: string;
	
	    static createFrom(source: any = {}) {
	        return new PendingPrompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeId = source["nodeId"];
	        this.inputId = source["inputId"];
	        this.prompt = source["prompt"];
	        this.shape = source["shape"];
	        this.options = source["options"];
	        this.round = source["round"];
	        this.createdAt = source["createdAt"];
	        this.promptId = source["promptId"];
	    }
	}
	export class Position {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}
	export class ProcessDef {
	    id: string;
	    name: string;
	    phase: string;
	    agentRole: string;
	    skillName: string;
	    description: string;
	    inputs: string[];
	    outputs: string[];
	    moduleId: string;
	    version: string;
	    mode?: string;
	    inputSpecs?: InputSpec[];
	    outputSpecs?: OutputSpec[];
	    gate?: IterationGate;
	
	    static createFrom(source: any = {}) {
	        return new ProcessDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phase = source["phase"];
	        this.agentRole = source["agentRole"];
	        this.skillName = source["skillName"];
	        this.description = source["description"];
	        this.inputs = source["inputs"];
	        this.outputs = source["outputs"];
	        this.moduleId = source["moduleId"];
	        this.version = source["version"];
	        this.mode = source["mode"];
	        this.inputSpecs = this.convertValues(source["inputSpecs"], InputSpec);
	        this.outputSpecs = this.convertValues(source["outputSpecs"], OutputSpec);
	        this.gate = this.convertValues(source["gate"], IterationGate);
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
	export class SprintStory {
	    id: string;
	    epicId: string;
	    status: string;
	    sequence: number;
	
	    static createFrom(source: any = {}) {
	        return new SprintStory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.epicId = source["epicId"];
	        this.status = source["status"];
	        this.sequence = source["sequence"];
	    }
	}
	export class SprintEpic {
	    id: string;
	    status: string;
	    stories: SprintStory[];
	
	    static createFrom(source: any = {}) {
	        return new SprintEpic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = source["status"];
	        this.stories = this.convertValues(source["stories"], SprintStory);
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
	export class SprintStatus {
	    generated: string;
	    lastUpdated: string;
	    project: string;
	    projectKey: string;
	    trackingSystem: string;
	    storyLocation: string;
	    epics: SprintEpic[];
	
	    static createFrom(source: any = {}) {
	        return new SprintStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generated = source["generated"];
	        this.lastUpdated = source["lastUpdated"];
	        this.project = source["project"];
	        this.projectKey = source["projectKey"];
	        this.trackingSystem = source["trackingSystem"];
	        this.storyLocation = source["storyLocation"];
	        this.epics = this.convertValues(source["epics"], SprintEpic);
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
	
	
	export class WorkflowEdge {
	    id: string;
	    source: string;
	    target: string;
	    sourceHandle?: string;
	    targetHandle?: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowEdge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.target = source["target"];
	        this.sourceHandle = source["sourceHandle"];
	        this.targetHandle = source["targetHandle"];
	    }
	}
	export class WorkflowNode {
	    id: string;
	    processId: string;
	    label: string;
	    position: Position;
	    status: string;
	    config: Record<string, string>;
	    tmuxTarget: string;
	    startedAt?: string;
	    storyId?: string;
	    nodeType?: string;
	    outputPaths?: Record<string, string>;
	    inputPaths?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.processId = source["processId"];
	        this.label = source["label"];
	        this.position = this.convertValues(source["position"], Position);
	        this.status = source["status"];
	        this.config = source["config"];
	        this.tmuxTarget = source["tmuxTarget"];
	        this.startedAt = source["startedAt"];
	        this.storyId = source["storyId"];
	        this.nodeType = source["nodeType"];
	        this.outputPaths = source["outputPaths"];
	        this.inputPaths = source["inputPaths"];
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
	export class WorkflowDef {
	    id: string;
	    name: string;
	    description: string;
	    repoPath?: string;
	    nodes: WorkflowNode[];
	    edges: WorkflowEdge[];
	    isTemplate: boolean;
	    templateId?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.repoPath = source["repoPath"];
	        this.nodes = this.convertValues(source["nodes"], WorkflowNode);
	        this.edges = this.convertValues(source["edges"], WorkflowEdge);
	        this.isTemplate = source["isTemplate"];
	        this.templateId = source["templateId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	
	export class WorkflowExecution {
	    id: string;
	    workflowId: string;
	    repoPath: string;
	    status: string;
	    nodes: WorkflowNode[];
	    startedAt: string;
	    currentNode: string;
	    nodeOutputs?: Record<string, string>;
	    nodeRounds?: Record<string, number>;
	    pendingPrompts?: PendingPrompt[];
	    nodeInputs?: Record<string, any>;
	    nodeInputHistory?: Record<string, Array<NodeInputEntry>>;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowExecution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workflowId = source["workflowId"];
	        this.repoPath = source["repoPath"];
	        this.status = source["status"];
	        this.nodes = this.convertValues(source["nodes"], WorkflowNode);
	        this.startedAt = source["startedAt"];
	        this.currentNode = source["currentNode"];
	        this.nodeOutputs = source["nodeOutputs"];
	        this.nodeRounds = source["nodeRounds"];
	        this.pendingPrompts = this.convertValues(source["pendingPrompts"], PendingPrompt);
	        this.nodeInputs = source["nodeInputs"];
	        this.nodeInputHistory = this.convertValues(source["nodeInputHistory"], Array<NodeInputEntry>, true);
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

}

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
	export class ModelInfo {
	    id: string;
	    alias: string;
	    displayName: string;
	    contextWindow: number;
	    tier: string;
	    isDefault: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.alias = source["alias"];
	        this.displayName = source["displayName"];
	        this.contextWindow = source["contextWindow"];
	        this.tier = source["tier"];
	        this.isDefault = source["isDefault"];
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
	    tokenSamples?: number[];
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
	        this.tokenSamples = source["tokenSamples"];
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
	export class TerminalSession {
	    sessionName: string;
	    paneTarget: string;
	    repoPath: string;
	    repoName: string;
	    sessionType: string;
	    model: string;
	    // Go type: time
	    spawnedAt: any;
	    isAlive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TerminalSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionName = source["sessionName"];
	        this.paneTarget = source["paneTarget"];
	        this.repoPath = source["repoPath"];
	        this.repoName = source["repoName"];
	        this.sessionType = source["sessionType"];
	        this.model = source["model"];
	        this.spawnedAt = this.convertValues(source["spawnedAt"], null);
	        this.isAlive = source["isAlive"];
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
	
	export class BranchInfo {
	    name: string;
	    current: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BranchInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.current = source["current"];
	    }
	}
	export class EditorSettings {
	    minimapEnabled: boolean;
	    wordWrap: string;
	    lineNumbers: string;
	    renderWhitespace: string;
	    tabSize: number;
	    insertSpaces: boolean;
	    cursorStyle: string;
	    cursorBlinking: string;
	    bracketPairColorization: boolean;
	    renderLineHighlight: string;
	    fontLigatures: boolean;
	    scrollBeyondLastLine: boolean;
	    smoothScrolling: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EditorSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minimapEnabled = source["minimapEnabled"];
	        this.wordWrap = source["wordWrap"];
	        this.lineNumbers = source["lineNumbers"];
	        this.renderWhitespace = source["renderWhitespace"];
	        this.tabSize = source["tabSize"];
	        this.insertSpaces = source["insertSpaces"];
	        this.cursorStyle = source["cursorStyle"];
	        this.cursorBlinking = source["cursorBlinking"];
	        this.bracketPairColorization = source["bracketPairColorization"];
	        this.renderLineHighlight = source["renderLineHighlight"];
	        this.fontLigatures = source["fontLigatures"];
	        this.scrollBeyondLastLine = source["scrollBeyondLastLine"];
	        this.smoothScrolling = source["smoothScrolling"];
	    }
	}
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
	export class RepoChoice {
	    name: string;
	    path: string;
	    branch: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.branch = source["branch"];
	    }
	}
	export class RepoStatusInfo {
	    dirty: boolean;
	    openPRs: number;
	    ahead: number;
	    behind: number;
	    protected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RepoStatusInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dirty = source["dirty"];
	        this.openPRs = source["openPRs"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.protected = source["protected"];
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
	    sidebarWidth?: number;
	    editorSettings?: EditorSettings;
	
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
	        this.sidebarWidth = source["sidebarWidth"];
	        this.editorSettings = this.convertValues(source["editorSettings"], EditorSettings);
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

}

