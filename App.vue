<script setup>
import {NMessageProvider,NDialogProvider,NModalProvider} from 'naive-ui'
import { NConfigProvider } from 'naive-ui'
import { zhCN, dateZhCN } from 'naive-ui'
</script>

<template>
	<n-config-provider :locale="zhCN" :date-locale="dateZhCN">
	<n-dialog-provider>
	<n-message-provider>
		<n-modal-provider>
		<router-view v-slot="{ Component, route }">
			<!-- <transition> -->
			<keep-alive include="Index">
				<component :is="Component" :key="route.path"></component>
			</keep-alive>
			<!-- </transition> -->
		</router-view>
		</n-modal-provider>
	</n-message-provider>
	</n-dialog-provider>
	</n-config-provider>
</template>

<style scoped></style>
