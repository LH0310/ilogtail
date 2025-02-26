/*
 * Copyright 2022 iLogtail Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "plugin/PluginRegistry.h"

#include <string>
#include <stdio.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include "logger/Logger.h"
#include "app_config/AppConfig.h"
#include "plugin/PluginCreator.h"
#include "plugin/StaticProcessorCreator.h"
#include "plugin/DynamicCProcessorCreator.h"

#include "plugin/CProcessor.h"
#include "processor/ProcessorSplitLogStringNative.h"
#include "processor/ProcessorSplitRegexNative.h"
#include "processor/ProcessorParseApsaraNative.h"
#include "processor/ProcessorParseDelimiterNative.h"
#include "processor/ProcessorParseJsonNative.h"
#include "processor/ProcessorParseRegexNative.h"
#include "processor/ProcessorParseTimestampNative.h"
#include "processor/ProcessorDesensitizeNative.h"
#include "processor/ProcessorFillGroupInfoNative.h"
#include "processor/ProcessorFilterNative.h"

namespace logtail {


PluginRegistry* PluginRegistry::GetInstance() {
    static PluginRegistry instance;
    return &instance;
}

void PluginRegistry::LoadPlugins() {
    LoadStaticPlugins();
    auto& plugins = AppConfig::GetInstance()->GetDynamicPlugins();
    LoadDynamicPlugins(plugins);
}

void PluginRegistry::UnloadPlugins() {
    // for (auto& kv : mPluginDict) {
    // if (node->plugin_type() == PLUGIN_TYPE_DYNAMIC) {
    //     CPluginRegistryItem* registry = reinterpret_cast<CPluginRegistryItem*>(node);
    //     if (strcmp(registry->mPlugin->language, "Go") == 0) {
    //         destroy_go_plugin_interface(registry->_handle,
    //         const_cast<plugin_interface_t*>(registry->mPlugin));
    //     }
    // }
    //     UnregisterCreator(kv.second.get());
    // }
    mPluginDict.clear();
}

std::unique_ptr<ProcessorInstance> PluginRegistry::CreateProcessor(const std::string& name,
                                                                   const std::string& pluginId) {
    return std::unique_ptr<ProcessorInstance>(
        static_cast<ProcessorInstance*>(Create(PROCESSOR_PLUGIN, name, pluginId).release()));
}

std::unique_ptr<PluginInstance>
PluginRegistry::Create(PluginCat cat, const std::string& name, const std::string& pluginId) {
    std::unique_ptr<PluginInstance> ins;
    auto creatorEntry = mPluginDict.find(PluginKey(cat, name));
    if (creatorEntry != mPluginDict.end()) {
        ins = creatorEntry->second->Create(pluginId);
    }
    return ins;
}

void PluginRegistry::LoadStaticPlugins() {
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorSplitLogStringNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorSplitRegexNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorParseApsaraNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorParseDelimiterNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorDesensitizeNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorParseJsonNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorParseRegexNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorParseTimestampNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorFillGroupInfoNative>());
    RegisterProcessorCreator(new StaticProcessorCreator<ProcessorFilterNative>());

    /* more native plugin registers here */
}

void PluginRegistry::LoadDynamicPlugins(const std::set<std::string>& plugins) {
    if (plugins.empty()) {
        return;
    }
    std::string error;
    auto pluginDir = AppConfig::GetInstance()->GetProcessExecutionDir() + "/plugins";
    for (auto& pluginName : plugins) {
        DynamicLibLoader loader;
        if (!loader.LoadDynLib(pluginName, error, pluginDir)) {
            LOG_ERROR(sLogger, ("open plugin", pluginName)("error", error));
            continue;
        }
        PluginCreator* creator = LoadProcessorPlugin(loader, pluginName);
        if (creator) {
            RegisterProcessorCreator(creator);
            continue;
        }
    }
}

PluginCreator* PluginRegistry::LoadProcessorPlugin(DynamicLibLoader& loader, const std::string pluginName) {
    std::string error;
    processor_interface_t* plugin = (processor_interface_t*)loader.LoadMethod("processor_interface", error);
    // if (!error.empty()) {
    //     loader.LoadMethod("x_cgo_init", error)
    //     if (error.empty()) { // try go plugin
    //         plugin = create_go_plugin_interface(handle);
    //     }
    // }
    if (!error.empty() || !plugin) {
        LOG_ERROR(sLogger, ("load method", "plugin_interface")("error", error));
        return nullptr;
    }
    if (plugin->version != PROCESSOR_INTERFACE_VERSION) {
        LOG_ERROR(sLogger,
                  ("load plugin", pluginName)("error", "plugin interface version mismatch")(
                      "expected", PROCESSOR_INTERFACE_VERSION)("actual", plugin->version));
        return nullptr;
    }
    return new DynamicCProcessorCreator(plugin, loader.Release());
}

void PluginRegistry::RegisterCreator(PluginCat cat, PluginCreator* creator) {
    if (!creator) {
        return;
    }
    mPluginDict.emplace(PluginKey(cat, creator->Name()), std::shared_ptr<PluginCreator>(creator));
}

void PluginRegistry::RegisterProcessorCreator(PluginCreator* creator) {
    RegisterCreator(PROCESSOR_PLUGIN, creator);
}

} // namespace logtail