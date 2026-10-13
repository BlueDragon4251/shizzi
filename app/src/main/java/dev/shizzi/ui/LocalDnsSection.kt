package dev.shizzi.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import dev.shizzi.LocalDnsRecord

/** Editable DNS records stored on the phone and applied live to the Shizuku datapath. */
@Composable
fun LocalDnsSection(
    enabled: Boolean,
    records: List<LocalDnsRecord>,
    onSetEnabled: (Boolean) -> Unit,
    onSave: (String?, LocalDnsRecord) -> Unit,
    onRemove: (String) -> Unit,
) {
    var editing by remember { mutableStateOf<LocalDnsRecord?>(null) }
    var showEditor by remember { mutableStateOf(false) }
    var name by remember { mutableStateOf("") }
    var ip by remember { mutableStateOf("") }
    var recordEnabled by remember { mutableStateOf(true) }

    fun edit(record: LocalDnsRecord?) {
        editing = record
        name = record?.name.orEmpty()
        ip = record?.ip.orEmpty()
        recordEnabled = record?.enabled ?: true
        showEditor = true
    }

    SettingsToggle(
        label = SettingsText("Local DNS", "Resolve custom hostnames for all hotspot clients."),
        isChecked = enabled,
        onCheckedChange = onSetEnabled,
    )

    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        records.forEach { record ->
            Column(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(record.name)
                        Text(record.ip)
                    }
                    TextButton(onClick = { edit(record) }) { Text("Edit") }
                }
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                ) {
                    TextButton(onClick = { onSave(record.name, record.copy(enabled = !record.enabled)) }) {
                        Text(if (record.enabled) "Disable" else "Enable")
                    }
                    TextButton(onClick = { onRemove(record.name) }) { Text("Delete") }
                }
            }
        }
        TextButton(onClick = { edit(null) }) { Text("+ Add DNS record") }
        if (records.isEmpty()) Text("Add a hostname and your Docker host's hotspot IPv4 address.")
    }

    if (showEditor) {
        val candidate = LocalDnsRecord(name, ip, recordEnabled).normalized()
        val duplicate = records.any { it.name == candidate.name && it.name != editing?.name }
        val canSave = candidate.isValid() && !duplicate
        AlertDialog(
            onDismissRequest = { showEditor = false },
            title = { Text(if (editing == null) "Add DNS record" else "Edit DNS record") },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    OutlinedTextField(
                        value = name, onValueChange = { name = it },
                        label = { Text("Hostname (e.g. photocraft)") },
                        singleLine = true,
                    )
                    OutlinedTextField(
                        value = ip, onValueChange = { ip = it },
                        label = { Text("IPv4 address (e.g. 10.201.244.112)") },
                        singleLine = true,
                    )
                    SettingsToggle(
                        label = SettingsText("Enabled"),
                        isChecked = recordEnabled,
                        onCheckedChange = { recordEnabled = it },
                    )
                    if (!canSave) Text("Enter a valid, unique hostname and IPv4 address.")
                }
            },
            confirmButton = {
                TextButton(
                    enabled = canSave,
                    onClick = {
                        onSave(editing?.name, candidate)
                        showEditor = false
                    },
                ) { Text("Save") }
            },
            dismissButton = {
                TextButton(onClick = { showEditor = false }) { Text("Cancel") }
            },
        )
    }
}
