package dev.shizzi

import org.json.JSONArray
import org.json.JSONObject
import java.util.Locale

data class LocalDnsRecord(val name: String, val ip: String, val enabled: Boolean = true) {
    fun normalized() = copy(name = name.trim().trimEnd('.').lowercase(Locale.ROOT), ip = ip.trim())

    fun isValid(): Boolean {
        val candidate = normalized()
        if (candidate.name.length !in 1..253) return false
        if (!candidate.name.split('.').all { label ->
                label.length in 1..63 && label.first() != '-' && label.last() != '-' &&
                    label.all { it in 'a'..'z' || it in '0'..'9' || it == '-' }
            }) return false
        val octets = candidate.ip.split('.')
        return octets.size == 4 && octets.all {
            it.isNotEmpty() && it.length <= 3 && it.all(Char::isDigit) &&
                (it.toIntOrNull()?.let { number -> number in 0..255 } == true)
        }
    }
}

internal fun parseLocalDnsRecords(json: String?): List<LocalDnsRecord> {
    if (json.isNullOrBlank()) return emptyList()
    return runCatching {
        val array = JSONArray(json)
        val records = mutableListOf<LocalDnsRecord>()
        for (i in 0 until array.length()) {
            val entry = array.getJSONObject(i)
            val record = LocalDnsRecord(
                entry.getString("name"), entry.getString("ip"), entry.optBoolean("enabled", true),
            ).normalized()
            if (record.isValid() && records.none { it.name == record.name }) records.add(record)
        }
        records
    }.getOrDefault(emptyList())
}

internal fun serializeLocalDnsRecords(records: List<LocalDnsRecord>): String =
    JSONArray().apply {
        records.forEach { record ->
            put(JSONObject().apply {
                put("name", record.name)
                put("ip", record.ip)
                put("enabled", record.enabled)
            })
        }
    }.toString()

internal fun Settings.localDnsConfigJson(): String = JSONObject().apply {
    put("enabled", localDnsEnabled)
    put("records", JSONArray(serializeLocalDnsRecords(localDnsRecords)))
}.toString()
