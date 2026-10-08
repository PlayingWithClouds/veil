package com.playingwithclouds.veil.ui.components

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.playingwithclouds.veil.ui.design.IconTap
import com.playingwithclouds.veil.ui.design.VeilIcons
import com.playingwithclouds.veil.ui.design.VeilTextField

/** A rounded search box with a clear button; [onSearch] fires on the keyboard's search key. */
@Composable
fun SearchField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    modifier: Modifier = Modifier,
    onSearch: (() -> Unit)? = null,
    focusRequester: FocusRequester? = null,
) {
    var fieldModifier = modifier.fillMaxWidth().padding(horizontal = 12.dp, vertical = 4.dp)
    if (focusRequester != null) {
        fieldModifier = fieldModifier.focusRequester(focusRequester)
    }
    VeilTextField(
        value = value,
        onValueChange = onValueChange,
        placeholder = placeholder,
        leadingIcon = VeilIcons.Search,
        trailing = {
            if (value.isNotEmpty()) {
                IconTap(VeilIcons.Close, contentDescription = "Clear", onClick = { onValueChange("") })
            }
        },
        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
        keyboardActions = KeyboardActions(onSearch = { onSearch?.invoke() }),
        modifier = fieldModifier,
    )
}
